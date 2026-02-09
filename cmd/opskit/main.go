package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	coreerr "opskit/internal/core/errors"
	"opskit/internal/core/executil"
	"opskit/internal/core/exitcode"
	"opskit/internal/core/fsx"
	"opskit/internal/core/lock"
	"opskit/internal/core/timeutil"
	"opskit/internal/engine"
	"opskit/internal/handover"
	"opskit/internal/installer"
	actionplugin "opskit/internal/plugins/actions"
	checkplugin "opskit/internal/plugins/checks"
	evidenceplugin "opskit/internal/plugins/evidence"
	"opskit/internal/schema"
	"opskit/internal/stages"
	"opskit/internal/state"
	"opskit/internal/templates"
	"opskit/internal/webserver"
)

const (
	statusJSONSchemaVersion = "v1"
	statusJSONCommand       = "opskit status"
	templateJSONSchemaVer   = "v1"
	templateJSONCommand     = "opskit template validate"
	templateListJSONCommand = "opskit template list"
)

type statusJSONPayload struct {
	Command       string                `json:"command"`
	ExitCode      int                   `json:"exitCode"`
	Health        string                `json:"health"`
	SchemaVersion string                `json:"schemaVersion"`
	GeneratedAt   string                `json:"generatedAt"`
	Overall       schema.OverallState   `json:"overall"`
	Lifecycle     schema.LifecycleState `json:"lifecycle"`
	Services      schema.ServicesState  `json:"services"`
	Artifacts     schema.ArtifactsState `json:"artifacts"`
}

type templateValidateIssue struct {
	Path    string `json:"path"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Advice  string `json:"advice,omitempty"`
}

type templateValidateJSONPayload struct {
	Command       string                  `json:"command"`
	SchemaVersion string                  `json:"schemaVersion"`
	Template      string                  `json:"template"`
	Valid         bool                    `json:"valid"`
	ErrorCount    int                     `json:"errorCount"`
	Issues        []templateValidateIssue `json:"issues"`
}

type templateListItem struct {
	Ref          string   `json:"ref"`
	Aliases      []string `json:"aliases,omitempty"`
	Source       string   `json:"source"`
	TemplateID   string   `json:"templateId"`
	Name         string   `json:"name"`
	Mode         string   `json:"mode"`
	ServiceScope string   `json:"serviceScope,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

type templateListJSONPayload struct {
	Command       string             `json:"command"`
	SchemaVersion string             `json:"schemaVersion"`
	Count         int                `json:"count"`
	Templates     []templateListItem `json:"templates"`
}

func main() {
	os.Exit(runCLI(os.Args[1:]))
}

func runCLI(args []string) int {
	if len(args) == 0 {
		printUsage()
		return exitcode.Precondition
	}

	switch args[0] {
	case "template":
		return cmdTemplate(args[1:])
	case "install":
		return cmdInstall(args[1:])
	case "run":
		return cmdRun(args[1:])
	case "status":
		return cmdStatus(args[1:])
	case "accept":
		return cmdAccept(args[1:])
	case "handover":
		return cmdHandover(args[1:])
	case "web":
		return cmdWeb(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		printUsage()
		return exitcode.Precondition
	}
}

func cmdTemplate(args []string) int {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: opskit template <validate|list> ...")
		return exitcode.Precondition
	}

	switch args[0] {
	case "validate":
		return cmdTemplateValidate(args[1:])
	case "list":
		return cmdTemplateList(args[1:])
	default:
		fmt.Fprintln(os.Stderr, "usage: opskit template <validate|list> ...")
		return exitcode.Precondition
	}
}

func cmdTemplateValidate(args []string) int {
	fs := flag.NewFlagSet("template validate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	varsRaw := fs.String("vars", "", "vars key=value[,key=value]")
	varsFile := fs.String("vars-file", "", "vars file (json or key=value lines)")
	output := fs.String("output", defaultOutputRoot(), "output root")
	jsonOutput := fs.Bool("json", false, "json output")
	if err := fs.Parse(normalizeTemplateValidateArgs(args)); err != nil {
		return exitcode.Precondition
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "missing template file or id")
		return exitcode.Precondition
	}
	ref := fs.Arg(0)
	t, _, err := templates.Resolve(templates.ResolveOptions{TemplateRef: ref, BaseDir: *output, VarsRaw: *varsRaw, VarsFile: *varsFile})
	if err != nil {
		issue := diagnoseTemplateValidateError(err)
		if *jsonOutput {
			printTemplateValidateJSON(ref, false, []templateValidateIssue{issue})
			return exitcode.Precondition
		}
		fmt.Fprintf(os.Stderr, "template invalid: %s\n", issue.Message)
		fmt.Fprintf(os.Stderr, "- path: %s\n", issue.Path)
		fmt.Fprintf(os.Stderr, "- code: %s\n", issue.Code)
		if strings.TrimSpace(issue.Advice) != "" {
			fmt.Fprintf(os.Stderr, "- advice: %s\n", issue.Advice)
		}
		return exitcode.Precondition
	}
	if err := validateTemplateContract(t); err != nil {
		issue := diagnoseTemplateValidateError(err)
		if *jsonOutput {
			printTemplateValidateJSON(ref, false, []templateValidateIssue{issue})
			return exitcode.Precondition
		}
		fmt.Fprintf(os.Stderr, "template invalid: %s\n", issue.Message)
		fmt.Fprintf(os.Stderr, "- path: %s\n", issue.Path)
		fmt.Fprintf(os.Stderr, "- code: %s\n", issue.Code)
		if strings.TrimSpace(issue.Advice) != "" {
			fmt.Fprintf(os.Stderr, "- advice: %s\n", issue.Advice)
		}
		return exitcode.Precondition
	}
	if *jsonOutput {
		printTemplateValidateJSON(ref, true, []templateValidateIssue{})
		return exitcode.Success
	}
	fmt.Printf("template valid: %s\n", ref)
	return exitcode.Success
}

func cmdTemplateList(args []string) int {
	fs := flag.NewFlagSet("template list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	jsonOutput := fs.Bool("json", false, "json output")
	if err := fs.Parse(args); err != nil {
		return exitcode.Precondition
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: opskit template list [--json]")
		return exitcode.Precondition
	}
	items, err := buildTemplateListItems(true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "template list failed: %v\n", err)
		return exitcode.Failure
	}
	if *jsonOutput {
		printTemplateListJSON(items)
		return exitcode.Success
	}
	fmt.Println("builtin templates:")
	for _, item := range items {
		fmt.Printf("- ref=%s id=%s mode=%s scope=%s source=%s\n", item.Ref, item.TemplateID, item.Mode, item.ServiceScope, item.Source)
		if len(item.Aliases) > 0 {
			fmt.Printf("  aliases=%s\n", strings.Join(item.Aliases, ","))
		}
		if len(item.Tags) > 0 {
			fmt.Printf("  tags=%s\n", strings.Join(item.Tags, ","))
		}
	}
	return exitcode.Success
}

func buildTemplateListItems(includeDemo bool) ([]templateListItem, error) {
	catalog, err := templates.Catalog(templates.CatalogOptions{IncludeDemo: includeDemo})
	if err != nil {
		return nil, err
	}
	items := make([]templateListItem, 0, len(catalog))
	for _, entry := range catalog {
		item := templateListItem{
			Ref:          entry.Ref,
			Aliases:      append([]string{}, entry.Aliases...),
			Source:       entry.Source,
			TemplateID:   entry.TemplateID,
			Name:         entry.Name,
			Mode:         entry.Mode,
			ServiceScope: entry.ServiceScope,
			Tags:         append([]string{}, entry.Tags...),
		}
		items = append(items, item)
	}
	return items, nil
}

func normalizeTemplateValidateArgs(raw []string) []string {
	if len(raw) == 0 {
		return raw
	}
	boolFlags := map[string]struct{}{
		"--json": {},
	}
	valueFlags := map[string]struct{}{
		"--vars":      {},
		"--vars-file": {},
		"--output":    {},
	}
	flags := make([]string, 0, len(raw))
	positionals := make([]string, 0, len(raw))
	for i := 0; i < len(raw); i++ {
		token := raw[i]
		name := token
		if eq := strings.Index(token, "="); eq >= 0 {
			name = token[:eq]
		}
		if _, ok := boolFlags[name]; ok {
			flags = append(flags, token)
			continue
		}
		if _, ok := valueFlags[name]; ok {
			flags = append(flags, token)
			if strings.Index(token, "=") < 0 && i+1 < len(raw) {
				flags = append(flags, raw[i+1])
				i++
			}
			continue
		}
		positionals = append(positionals, token)
	}
	return append(flags, positionals...)
}

func printTemplateValidateJSON(ref string, valid bool, issues []templateValidateIssue) {
	payload := templateValidateJSONPayload{
		Command:       templateJSONCommand,
		SchemaVersion: templateJSONSchemaVer,
		Template:      ref,
		Valid:         valid,
		ErrorCount:    len(issues),
		Issues:        issues,
	}
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "template validate json marshal failed: %v\n", err)
		return
	}
	fmt.Println(string(body))
}

func printTemplateListJSON(items []templateListItem) {
	payload := templateListJSONPayload{
		Command:       templateListJSONCommand,
		SchemaVersion: templateJSONSchemaVer,
		Count:         len(items),
		Templates:     items,
	}
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "template list json marshal failed: %v\n", err)
		return
	}
	fmt.Println(string(body))
}

func diagnoseTemplateValidateError(err error) templateValidateIssue {
	msg := strings.TrimSpace(err.Error())
	issue := templateValidateIssue{
		Path:    "template",
		Code:    "template_invalid",
		Message: msg,
		Advice:  "check template file and vars, then rerun `opskit template validate`",
	}

	switch {
	case strings.Contains(msg, "unknown template id:"):
		issue.Path = "template.ref"
		issue.Code = "template_unknown_id"
		issue.Advice = "run `opskit template list` for builtin/demo refs, or pass a valid .json file path"
		return issue
	case strings.HasPrefix(msg, "vars-file "):
		issue.Path = "vars-file"
		issue.Code = "vars_file_invalid"
		issue.Advice = "use JSON object or key=value lines; comments are allowed with # or //"
		return issue
	case strings.Contains(msg, "no such file or directory"):
		issue.Path = "template.file"
		issue.Code = "template_file_not_found"
		issue.Advice = "check template path and file permissions"
		return issue
	case strings.Contains(msg, "permission denied"):
		issue.Path = "template.file"
		issue.Code = "template_file_permission_denied"
		issue.Advice = "ensure read permission on template/vars files"
		return issue
	case strings.Contains(msg, "template: json: unknown field "):
		field := between(msg, `unknown field "`, `"`)
		if field != "" {
			issue.Path = "template." + field
			issue.Message = "unknown template JSON field: " + field
		}
		issue.Code = "template_unknown_field"
		issue.Advice = "remove unsupported field or update template schema"
		return issue
	case strings.Contains(msg, "template: unexpected extra JSON content"):
		issue.Path = "template"
		issue.Code = "template_json_trailing_content"
		issue.Advice = "ensure the template file contains exactly one JSON object"
		return issue
	case strings.Contains(msg, "unresolved var "):
		path, token := splitUnresolvedVar(msg)
		if path != "" {
			issue.Path = path
		}
		issue.Code = "template_unresolved_var"
		issue.Advice = "define " + token + " in template.vars and pass value via --vars/--vars-file"
		return issue
	case strings.Contains(msg, "invalid stage id"):
		issue.Path = "template.stages"
		issue.Code = "template_stage_invalid"
		issue.Advice = "stage id must be one of A,B,C,D,E,F"
		return issue
	case strings.Contains(msg, ".params.severity"):
		issue.Path = extractTemplatePath(msg)
		issue.Code = "template_severity_invalid"
		issue.Advice = "severity must be one of info, warn, fail"
		return issue
	case strings.Contains(msg, ".kind unsupported:") || strings.Contains(msg, "plugin not found:"):
		issue.Path = extractTemplatePath(msg)
		issue.Code = "template_plugin_kind_unsupported"
		issue.Advice = "use supported check/action/evidence kinds or register the plugin before running"
		return issue
	}

	path := extractTemplatePath(msg)
	if path != "" {
		issue.Path = path
	}
	if strings.Contains(msg, "template.vars.") {
		issue.Code = "template_var_invalid"
		issue.Advice = "check var type/enum/default or pass required value via --vars/--vars-file"
		if strings.Contains(msg, " is required") {
			issue.Code = "template_var_required"
			issue.Advice = "pass the required var via --vars or --vars-file"
		} else if strings.Contains(msg, "expects ") {
			issue.Code = "template_var_type_mismatch"
			issue.Advice = typeMismatchAdvice(msg)
		} else if strings.Contains(msg, "invalid value") || strings.Contains(msg, "default not in enum") {
			issue.Code = "template_var_enum_mismatch"
			issue.Advice = "use one of allowed enum values declared in template.vars"
		}
	}
	return issue
}

func typeMismatchAdvice(msg string) string {
	switch {
	case strings.Contains(msg, "expects json array"):
		return "pass valid JSON array via --vars-file (recommended), e.g. [80,443]"
	case strings.Contains(msg, "expects json object"):
		return "pass valid JSON object via --vars-file (recommended), e.g. {\"key\":\"value\"}"
	case strings.Contains(msg, "expects bool"):
		return "use true/false (or 1/0 when supported by bool parser)"
	case strings.Contains(msg, "expects int"):
		return "use integer value, e.g. 18080"
	case strings.Contains(msg, "expects number"):
		return "use numeric value, e.g. 0.75"
	default:
		return "fix var value type (int/number/bool/json array/json object)"
	}
}

func extractTemplatePath(msg string) string {
	start := strings.Index(msg, "template")
	if start < 0 {
		return ""
	}
	segment := msg[start:]
	stop := len(segment)
	for i, r := range segment {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '[' || r == ']' {
			continue
		}
		stop = i
		break
	}
	path := strings.TrimSpace(segment[:stop])
	if path != "" {
		return path
	}
	return ""
}

func splitUnresolvedVar(msg string) (string, string) {
	idx := strings.Index(msg, ": unresolved var ")
	if idx < 0 {
		return "", "${VAR}"
	}
	path := strings.TrimSpace(msg[:idx])
	token := strings.TrimSpace(msg[idx+len(": unresolved var "):])
	if token == "" {
		token = "${VAR}"
	}
	return path, token
}

func between(s, left, right string) string {
	start := strings.Index(s, left)
	if start < 0 {
		return ""
	}
	start += len(left)
	end := strings.Index(s[start:], right)
	if end < 0 {
		return ""
	}
	return s[start : start+end]
}

func cmdInstall(args []string) int {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	templateRef := fs.String("template", "generic-manage-v1", "template id or path")
	varsRaw := fs.String("vars", "", "vars key=value[,key=value]")
	varsFile := fs.String("vars-file", "", "vars file (json or key=value lines)")
	dryRun := fs.Bool("dry-run", false, "dry run")
	fix := fs.Bool("fix", false, "include disabled template steps")
	listenAddr := fs.String("listen", ":18080", "web listen address")
	systemdDir := fs.String("systemd-dir", "", "systemd unit directory (default: /etc/systemd/system when output=/var/lib/opskit, else <output>/systemd)")
	binaryPath := fs.String("binary-path", "", "opskit binary path used in generated systemd units")
	noSystemd := fs.Bool("no-systemd", false, "skip systemd unit generation")
	output := fs.String("output", defaultOutputRoot(), "output root")
	if err := fs.Parse(args); err != nil {
		return exitcode.Precondition
	}

	t, _, err := templates.Resolve(templates.ResolveOptions{TemplateRef: *templateRef, BaseDir: *output, VarsRaw: *varsRaw, VarsFile: *varsFile})
	if err != nil {
		fmt.Fprintf(os.Stderr, "install precondition failed: %v\n", err)
		return exitcode.Precondition
	}
	if err := ensureTemplateHasStages(t, []string{"A", "D"}); err != nil {
		fmt.Fprintf(os.Stderr, "install precondition failed: %v\n", err)
		return exitcode.Precondition
	}
	if err := validateTemplateContract(t); err != nil {
		fmt.Fprintf(os.Stderr, "install precondition failed: %v\n", err)
		return exitcode.Precondition
	}

	store := state.NewStore(state.NewPaths(*output))
	res, err := installer.InstallLayout(store, *dryRun)
	if err != nil {
		fmt.Fprintf(os.Stderr, "install failed: %v\n", err)
		return exitcode.Failure
	}
	if !*noSystemd {
		binPath := strings.TrimSpace(*binaryPath)
		if binPath == "" {
			detected, binErr := os.Executable()
			if binErr != nil || strings.TrimSpace(detected) == "" {
				detected = os.Args[0]
			}
			binPath = detected
		}
		dir := *systemdDir
		if strings.TrimSpace(dir) == "" {
			dir = defaultSystemdDir(*output)
		}
		sysRes, sysErr := installer.InstallSystemdUnits(store.Paths(), installer.SystemdOptions{
			SystemdDir: dir,
			BinaryPath: binPath,
			ListenAddr: *listenAddr,
		}, *dryRun)
		if sysErr != nil {
			fmt.Fprintf(os.Stderr, "install systemd failed: %v\n", sysErr)
			return exitcode.Failure
		}
		res.Actions = append(res.Actions, sysRes.Actions...)
		if filepath.Clean(dir) == "/etc/systemd/system" {
			activateRes, activateErr := installer.ActivateSystemdUnits(context.Background(), executil.SystemRunner{}, *dryRun)
			if activateErr != nil {
				if errors.Is(activateErr, coreerr.ErrPreconditionFailed) {
					fmt.Fprintf(os.Stderr, "install precondition failed: %v\n", activateErr)
					return exitcode.Precondition
				}
				fmt.Fprintf(os.Stderr, "install systemd activation failed: %v\n", activateErr)
				return exitcode.Failure
			}
			res.Actions = append(res.Actions, activateRes.Actions...)
		}
	}

	if *dryRun {
		fmt.Println("install dry-run actions:")
		for _, action := range res.Actions {
			fmt.Printf("- %s\n", action)
		}
		fmt.Println("bootstrap plan: run A,D")
		return exitcode.Success
	}

	unlock, code := acquireGlobalLock(store)
	if code != exitcode.Success {
		return code
	}
	defer unlock()

	if err := store.InitStateIfMissing(t.ID); err != nil {
		fmt.Fprintf(os.Stderr, "state init failed: %v\n", err)
		return exitcode.Failure
	}

	runCode, err := executeStages(context.Background(), store, t, []string{"A", "D"}, false, *fix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bootstrap failed: %v\n", err)
		return mapErrorToExit(err)
	}

	fmt.Printf("install completed: output=%s ui=%s\n", filepath.Clean(*output), filepath.Join(store.Paths().UIDir, "index.html"))
	return runCode
}

func cmdRun(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: opskit run <A|B|C|D|E|F|AF> [--template ...] [--vars ...] [--dry-run] [--output ...]")
		return exitcode.Precondition
	}

	stageSelector := args[0]
	selected, err := engine.SelectStages(stageSelector)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return exitcode.Precondition
	}

	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	templateRef := fs.String("template", "", "template id or path (default: active template from state, fallback generic-manage-v1)")
	varsRaw := fs.String("vars", "", "vars key=value[,key=value]")
	varsFile := fs.String("vars-file", "", "vars file (json or key=value lines)")
	dryRun := fs.Bool("dry-run", false, "dry run")
	fix := fs.Bool("fix", false, "include disabled template steps")
	output := fs.String("output", defaultOutputRoot(), "output root")
	if err := fs.Parse(args[1:]); err != nil {
		return exitcode.Precondition
	}

	selectedTemplate := resolveTemplateRef(*templateRef, *output)
	t, _, err := templates.Resolve(templates.ResolveOptions{TemplateRef: selectedTemplate, BaseDir: *output, VarsRaw: *varsRaw, VarsFile: *varsFile})
	if err != nil {
		fmt.Fprintf(os.Stderr, "run precondition failed: %v\n", err)
		return exitcode.Precondition
	}
	if err := ensureTemplateHasStages(t, selected); err != nil {
		fmt.Fprintf(os.Stderr, "run precondition failed: %v\n", err)
		return exitcode.Precondition
	}
	if err := validateTemplateContract(t); err != nil {
		fmt.Fprintf(os.Stderr, "run precondition failed: %v\n", err)
		return exitcode.Precondition
	}

	store := state.NewStore(state.NewPaths(*output))
	if *dryRun {
		plan := engine.BuildPlan(t, selected, *fix)
		fmt.Printf("run dry-run template=%s stages=%s\n", t.ID, strings.Join(selected, ","))
		if *fix {
			fmt.Println("mode: include disabled steps (--fix)")
		}
		for _, s := range plan.Stages {
			fmt.Printf("- stage %s checks=%d actions=%d evidence=%d\n", s.StageID, len(s.Checks), len(s.Actions), len(s.Evidence))
		}
		return exitcode.Success
	}

	unlock, code := acquireGlobalLock(store)
	if code != exitcode.Success {
		return code
	}
	defer unlock()

	runCode, err := executeStages(context.Background(), store, t, selected, false, *fix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
		return mapErrorToExit(err)
	}
	return runCode
}

func cmdAccept(args []string) int {
	fs := flag.NewFlagSet("accept", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	templateRef := fs.String("template", "", "template id or path (default: active template from state, fallback generic-manage-v1)")
	varsRaw := fs.String("vars", "", "vars key=value[,key=value]")
	varsFile := fs.String("vars-file", "", "vars file (json or key=value lines)")
	dryRun := fs.Bool("dry-run", false, "dry run")
	fix := fs.Bool("fix", false, "include disabled template steps")
	output := fs.String("output", defaultOutputRoot(), "output root")
	if err := fs.Parse(args); err != nil {
		return exitcode.Precondition
	}
	selectedTemplate := resolveTemplateRef(*templateRef, *output)
	t, _, err := templates.Resolve(templates.ResolveOptions{TemplateRef: selectedTemplate, BaseDir: *output, VarsRaw: *varsRaw, VarsFile: *varsFile})
	if err != nil {
		fmt.Fprintf(os.Stderr, "accept precondition failed: %v\n", err)
		return exitcode.Precondition
	}
	if err := ensureTemplateHasStages(t, []string{"F"}); err != nil {
		fmt.Fprintf(os.Stderr, "accept precondition failed: %v\n", err)
		return exitcode.Precondition
	}
	if err := validateTemplateContract(t); err != nil {
		fmt.Fprintf(os.Stderr, "accept precondition failed: %v\n", err)
		return exitcode.Precondition
	}

	store := state.NewStore(state.NewPaths(*output))
	if *dryRun {
		plan := engine.BuildPlan(t, []string{"F"}, *fix)
		fmt.Printf("accept dry-run template=%s\n", t.ID)
		for _, s := range plan.Stages {
			fmt.Printf("- stage %s checks=%d actions=%d evidence=%d\n", s.StageID, len(s.Checks), len(s.Actions), len(s.Evidence))
		}
		return exitcode.Success
	}
	unlock, code := acquireGlobalLock(store)
	if code != exitcode.Success {
		return code
	}
	defer unlock()
	runCode, err := executeStages(context.Background(), store, t, []string{"F"}, false, *fix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "accept failed: %v\n", err)
		return mapErrorToExit(err)
	}
	return runCode
}

func cmdHandover(args []string) int {
	fs := flag.NewFlagSet("handover", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	output := fs.String("output", defaultOutputRoot(), "output root")
	if err := fs.Parse(args); err != nil {
		return exitcode.Precondition
	}

	store := state.NewStore(state.NewPaths(*output))
	unlock, code := acquireGlobalLock(store)
	if code != exitcode.Success {
		return code
	}
	defer unlock()
	if err := store.InitStateIfMissing(""); err != nil {
		fmt.Fprintf(os.Stderr, "handover init failed: %v\n", err)
		return exitcode.Failure
	}
	artifacts, err := store.ReadArtifacts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "handover failed: %v\n", err)
		return exitcode.Failure
	}

	res, err := handover.Generate(store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "handover generation failed: %v\n", err)
		return exitcode.Failure
	}
	artifacts.Reports = append(artifacts.Reports, res.ReportHTML, res.ReportJSON)
	artifacts.Bundles = append(artifacts.Bundles, res.Bundle)
	if err := store.WriteArtifacts(artifacts); err != nil {
		fmt.Fprintf(os.Stderr, "handover artifacts failed: %v\n", err)
		return exitcode.Failure
	}
	fmt.Printf("handover generated: %s, %s, %s\n", res.ReportHTML.Path, res.ReportJSON.Path, res.Bundle.Path)
	return exitcode.Success
}

func cmdStatus(args []string) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	output := fs.String("output", defaultOutputRoot(), "output root")
	jsonOutput := fs.Bool("json", false, "print status as json")
	if err := fs.Parse(args); err != nil {
		return exitcode.Precondition
	}

	store := state.NewStore(state.NewPaths(*output))
	unlock, code := acquireGlobalLock(store)
	if code != exitcode.Success {
		return code
	}
	defer unlock()

	if err := store.InitStateIfMissing(""); err != nil {
		fmt.Fprintf(os.Stderr, "status init failed: %v\n", err)
		return exitcode.Failure
	}

	overall, err := store.ReadOverall()
	if err != nil {
		fmt.Fprintf(os.Stderr, "status unavailable: %v\n", err)
		return exitcode.Failure
	}
	lifecycle, err := store.ReadLifecycle()
	if err != nil {
		fmt.Fprintf(os.Stderr, "status unavailable: %v\n", err)
		return exitcode.Failure
	}
	services, err := store.ReadServices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "status unavailable: %v\n", err)
		return exitcode.Failure
	}
	artifacts, err := store.ReadArtifacts()
	if err != nil {
		fmt.Fprintf(os.Stderr, "status unavailable: %v\n", err)
		return exitcode.Failure
	}

	s, issues := state.DeriveOverall(lifecycle)
	overall.OverallStatus = s
	overall.OpenIssuesCount = issues
	overall.LastRefreshTime = timeutil.NowISO8601()
	overall.RecoverSummary = state.DeriveRecoverSummary(store.Paths(), lifecycle)
	if err := store.WriteLifecycle(lifecycle); err != nil {
		fmt.Fprintf(os.Stderr, "status write failed: %v\n", err)
		return exitcode.Failure
	}
	if err := store.WriteServices(services); err != nil {
		fmt.Fprintf(os.Stderr, "status write failed: %v\n", err)
		return exitcode.Failure
	}
	if err := store.WriteArtifacts(artifacts); err != nil {
		fmt.Fprintf(os.Stderr, "status write failed: %v\n", err)
		return exitcode.Failure
	}
	if err := store.WriteOverall(overall); err != nil {
		fmt.Fprintf(os.Stderr, "status write failed: %v\n", err)
		return exitcode.Failure
	}
	if err := writeTemplateCatalogState(store.Paths()); err != nil {
		fmt.Fprintf(os.Stderr, "status template catalog warning: %v\n", err)
	}
	finalCode := exitForLifecycle(lifecycle)

	if *jsonOutput {
		payload := buildStatusJSONPayload(overall, lifecycle, services, artifacts, finalCode)
		body, err := json.MarshalIndent(payload, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "status marshal failed: %v\n", err)
			return exitcode.Failure
		}
		fmt.Println(string(body))
		return finalCode
	}

	fmt.Printf("overall=%s issues=%d template=%s refreshed=%s\n", overall.OverallStatus, overall.OpenIssuesCount, strings.Join(overall.ActiveTemplates, ","), overall.LastRefreshTime)
	if overall.RecoverSummary != nil {
		rs := overall.RecoverSummary
		fmt.Printf("recover last=%s trigger=%s ok=%d fail=%d warn=%d circuitOpen=%t cooldownUntil=%s\n",
			rs.LastStatus, rs.LastTrigger, rs.SuccessCount, rs.FailureCount, rs.WarnCount, rs.CircuitOpen, rs.CooldownUntil)
	}
	for _, s := range lifecycle.Stages {
		fmt.Printf("- %s %-16s %s\n", s.StageID, s.Name, s.Status)
	}
	return finalCode
}

func writeTemplateCatalogState(paths state.Paths) error {
	items, err := buildTemplateListItems(true)
	if err != nil {
		return err
	}
	payload := templateListJSONPayload{
		Command:       templateListJSONCommand,
		SchemaVersion: templateJSONSchemaVer,
		Count:         len(items),
		Templates:     items,
	}
	return fsx.AtomicWriteJSON(filepath.Join(paths.StateDir, "templates.json"), payload)
}

func buildStatusJSONPayload(overall schema.OverallState, lifecycle schema.LifecycleState, services schema.ServicesState, artifacts schema.ArtifactsState, exitCode int) statusJSONPayload {
	generatedAt := strings.TrimSpace(overall.LastRefreshTime)
	if generatedAt == "" {
		generatedAt = timeutil.NowISO8601()
	}
	return statusJSONPayload{
		Command:       statusJSONCommand,
		ExitCode:      exitCode,
		Health:        statusHealth(exitCode),
		SchemaVersion: statusJSONSchemaVersion,
		GeneratedAt:   generatedAt,
		Overall:       overall,
		Lifecycle:     lifecycle,
		Services:      services,
		Artifacts:     artifacts,
	}
}

func statusHealth(exitCode int) string {
	switch exitCode {
	case exitcode.Success:
		return "ok"
	case exitcode.PartialSuccess:
		return "warn"
	case exitcode.Failure:
		return "fail"
	default:
		return "unknown"
	}
}

func cmdWeb(args []string) int {
	fs := flag.NewFlagSet("web", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	output := fs.String("output", defaultOutputRoot(), "output root")
	listenAddr := fs.String("listen", ":18080", "web listen address")
	if err := fs.Parse(args); err != nil {
		return exitcode.Precondition
	}
	store := state.NewStore(state.NewPaths(*output))
	if err := store.EnsureLayout(); err != nil {
		fmt.Fprintf(os.Stderr, "web layout failed: %v\n", err)
		return exitcode.Failure
	}
	if err := webserver.Serve(store.Paths(), *listenAddr); err != nil {
		fmt.Fprintf(os.Stderr, "web server failed: %v\n", err)
		return exitcode.Failure
	}
	return exitcode.Success
}

func executeStages(ctx context.Context, store *state.Store, t schema.Template, selected []string, dryRun bool, includeDisabled bool) (int, error) {
	checkReg := checkplugin.NewRegistry()
	checkplugin.RegisterBuiltins(checkReg)
	actionReg := actionplugin.NewRegistry()
	actionplugin.RegisterBuiltins(actionReg)
	evidenceReg := evidenceplugin.NewRegistry()
	evidenceplugin.RegisterBuiltins(evidenceReg)

	if err := ensureTemplateHasStages(t, selected); err != nil {
		return exitcode.Precondition, err
	}

	rt := &engine.Runtime{
		Store:            store,
		CheckRegistry:    checkReg,
		ActionRegistry:   actionReg,
		EvidenceRegistry: evidenceReg,
		Exec:             executil.SystemRunner{},
		Plan:             engine.BuildPlan(t, selected, includeDisabled),
		Options: engine.RunOptions{
			TemplateID:     t.ID,
			TemplateMode:   t.Mode,
			SelectedStages: selected,
			DryRun:         dryRun,
		},
	}
	if err := engine.ValidatePlanPlugins(rt.Plan, checkReg, actionReg, evidenceReg); err != nil {
		return exitcode.Precondition, err
	}
	runner := engine.NewRunner(stages.DefaultExecutors())
	results, err := runner.Execute(ctx, rt)
	if err != nil {
		return exitcode.Failure, err
	}

	status := stageResultsExit(results)
	for _, r := range results {
		fmt.Printf("stage %s -> %s\n", r.StageID, r.Status)
	}
	return status, nil
}

func acquireGlobalLock(store *state.Store) (func(), int) {
	l, err := lock.Acquire(store.Paths().LockFile)
	if err != nil {
		if errors.Is(err, coreerr.ErrLocked) {
			fmt.Fprintln(os.Stderr, "another opskit operation is running")
			return func() {}, exitcode.ManualIntervention
		}
		fmt.Fprintf(os.Stderr, "lock acquire failed: %v\n", err)
		return func() {}, exitcode.Failure
	}
	return func() { _ = l.Release() }, exitcode.Success
}

func stageResultsExit(results []engine.StageResult) int {
	hasWarn := false
	for _, r := range results {
		if r.Status == schema.StatusFailed {
			return exitcode.Failure
		}
		if r.Status == schema.StatusWarn {
			hasWarn = true
		}
	}
	if hasWarn {
		return exitcode.PartialSuccess
	}
	return exitcode.Success
}

func exitForLifecycle(lifecycle schema.LifecycleState) int {
	hasWarn := false
	for _, s := range lifecycle.Stages {
		if s.Status == schema.StatusFailed {
			return exitcode.Failure
		}
		if s.Status == schema.StatusWarn {
			hasWarn = true
		}
	}
	if hasWarn {
		return exitcode.PartialSuccess
	}
	return exitcode.Success
}

func mapErrorToExit(err error) int {
	if errors.Is(err, coreerr.ErrPreconditionFailed) {
		return exitcode.Precondition
	}
	if errors.Is(err, coreerr.ErrPartialSuccess) {
		return exitcode.PartialSuccess
	}
	if errors.Is(err, coreerr.ErrLocked) {
		return exitcode.ManualIntervention
	}
	return exitcode.Failure
}

func ensureTemplateHasStages(t schema.Template, selected []string) error {
	if len(selected) == 0 {
		return nil
	}
	missing := make([]string, 0, len(selected))
	for _, id := range selected {
		if _, ok := t.Stages[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("%w: template %s missing required stages: %s", coreerr.ErrPreconditionFailed, t.ID, strings.Join(missing, ","))
}

func validateTemplateContract(t schema.Template) error {
	checkReg := checkplugin.NewRegistry()
	checkplugin.RegisterBuiltins(checkReg)
	actionReg := actionplugin.NewRegistry()
	actionplugin.RegisterBuiltins(actionReg)
	evidenceReg := evidenceplugin.NewRegistry()
	evidenceplugin.RegisterBuiltins(evidenceReg)
	plan := templates.BuildPlanWithOptions(t, nil, true)
	return engine.ValidatePlanPlugins(plan, checkReg, actionReg, evidenceReg)
}

func defaultOutputRoot() string {
	if v := os.Getenv("OPSKIT_OUTPUT"); strings.TrimSpace(v) != "" {
		return v
	}
	return "/var/lib/opskit"
}

func defaultSystemdDir(output string) string {
	if filepath.Clean(output) == "/var/lib/opskit" {
		return "/etc/systemd/system"
	}
	return filepath.Join(output, "systemd")
}

func resolveTemplateRef(rawTemplateRef, output string) string {
	if strings.TrimSpace(rawTemplateRef) != "" {
		return rawTemplateRef
	}
	store := state.NewStore(state.NewPaths(output))
	overall, err := store.ReadOverall()
	if err == nil {
		for _, t := range overall.ActiveTemplates {
			if strings.TrimSpace(t) != "" {
				return t
			}
		}
	}
	return "generic-manage-v1"
}

func printUsage() {
	fmt.Println("OpsKit v1")
	fmt.Println("usage:")
	fmt.Println("  opskit install [--template id|path] [--vars k=v] [--vars-file file] [--dry-run] [--fix] [--output dir] [--systemd-dir dir] [--binary-path /path/opskit]")
	fmt.Println("  opskit run <A|B|C|D|E|F|AF> [--template id|path] [--vars k=v] [--vars-file file] [--dry-run] [--fix] [--output dir]")
	fmt.Println("  opskit status [--output dir] [--json]")
	fmt.Println("  opskit accept [--template id|path] [--vars k=v] [--vars-file file] [--dry-run] [--fix] [--output dir]")
	fmt.Println("  opskit handover [--output dir]")
	fmt.Println("  opskit web [--output dir] [--listen :18080]")
	fmt.Println("  opskit template validate <file|id> [--vars k=v] [--vars-file file] [--output dir] [--json]")
	fmt.Println("  opskit template list [--json]")
}
