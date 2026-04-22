package config

// Merge applies overlay onto base. Scalars in overlay replace base; pointer
// bools in overlay replace base when non-nil; string lists are a deduplicated
// union (base first, then overlay items not already present).
//
// Merge does NOT enforce lockdown. Use MergeRespectingLockdown for the full
// precedence chain (global -> project -> env) with lockdown enforcement.
func Merge(base, overlay Config) Config {
	out := base
	if overlay.Version != 0 {
		out.Version = overlay.Version
	}
	if overlay.Mode != "" {
		out.Mode = overlay.Mode
	}
	if overlay.Guard.Level != "" {
		out.Guard.Level = overlay.Guard.Level
	}
	if overlay.Guard.Lockdown {
		out.Guard.Lockdown = true
	}
	out.Guard.Tools = mergeTools(out.Guard.Tools, overlay.Guard.Tools)
	out.Guard.AdditionalDirs = unionStrings(out.Guard.AdditionalDirs, overlay.Guard.AdditionalDirs)
	out.Guard.BlockedCommands = unionStrings(out.Guard.BlockedCommands, overlay.Guard.BlockedCommands)
	out.Guard.AllowedNetwork = unionStrings(out.Guard.AllowedNetwork, overlay.Guard.AllowedNetwork)
	out.Guard.SensitivePaths = unionStrings(out.Guard.SensitivePaths, overlay.Guard.SensitivePaths)
	if overlay.Telemetry.EasterEgg {
		out.Telemetry.EasterEgg = true
	}
	return out
}

// MergeRespectingLockdown chains global -> project -> env and enforces lockdown:
// if global.Guard.Lockdown is true, any attempt by project or env to loosen
// guard.level is dropped and recorded as a violation.
//
// Returns the merged config plus a list of human-readable violation messages
// (empty when nothing was dropped).
func MergeRespectingLockdown(global, project, env Config) (Config, []string) {
	var violations []string
	baseline := global.Guard.Level
	locked := global.Guard.Lockdown

	merged := Merge(Config{}, global)

	if locked && project.Guard.Level != "" && LevelRank(project.Guard.Level) < LevelRank(baseline) {
		violations = append(violations, "project config tried to loosen guard.level from "+baseline+" to "+project.Guard.Level+" but lockdown is set; keeping "+baseline)
		project.Guard.Level = ""
	}
	merged = Merge(merged, project)

	if locked && env.Guard.Level != "" && LevelRank(env.Guard.Level) < LevelRank(baseline) {
		violations = append(violations, "IFLY_GUARD env tried to loosen guard.level from "+baseline+" to "+env.Guard.Level+" but lockdown is set; keeping "+baseline)
		env.Guard.Level = ""
	}
	merged = Merge(merged, env)

	return merged, violations
}

func mergeTools(base, over Tools) Tools {
	out := base
	if over.Bash != nil {
		out.Bash = over.Bash
	}
	if over.Edit != nil {
		out.Edit = over.Edit
	}
	if over.Write != nil {
		out.Write = over.Write
	}
	if over.MultiEdit != nil {
		out.MultiEdit = over.MultiEdit
	}
	if over.NotebookEdit != nil {
		out.NotebookEdit = over.NotebookEdit
	}
	if over.Read != nil {
		out.Read = over.Read
	}
	if over.Glob != nil {
		out.Glob = over.Glob
	}
	if over.Grep != nil {
		out.Grep = over.Grep
	}
	if over.WebFetch != nil {
		out.WebFetch = over.WebFetch
	}
	if over.WebSearch != nil {
		out.WebSearch = over.WebSearch
	}
	return out
}

func unionStrings(a, b []string) []string {
	seen := make(map[string]struct{}, len(a)+len(b))
	out := make([]string, 0, len(a)+len(b))
	for _, v := range a {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	for _, v := range b {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	return out
}
