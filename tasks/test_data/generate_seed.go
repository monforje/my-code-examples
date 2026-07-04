//go:build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type task struct {
	file        string
	taskName    string
	title       string
	description string
	taskType    string
	level       string
	tags        []string
	languages   []string
}

func main() {
	backendDir := "test_data/spec_markdown/backend"
	frontendDir := "test_data/spec_markdown/frontend"

	var tasks []task

	backendFiles := readDir(backendDir)
	for _, f := range backendFiles {
		content := readFile(filepath.Join(backendDir, f))
		title := extractTitle(content)
		description := extractDescription(content)

		t := task{file: f, taskName: taskNameFromFile(f), title: title, description: description, taskType: "backend"}

		switch {
		case strings.Contains(f, "01_auth"):
			t.level = "middle"
			t.tags = []string{"Security", "JWT", "REST"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "02_session"):
			t.level = "middle"
			t.tags = []string{"Security", "JWT", "REST", "Testing"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "03_password_recovery"):
			t.level = "middle"
			t.tags = []string{"Security", "REST", "Testing"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "04_email_change"):
			t.level = "middle"
			t.tags = []string{"Security", "REST", "Microservices"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "05_account_delete"):
			t.level = "middle"
			t.tags = []string{"Security", "REST", "Testing"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "06_users_event_sync"):
			t.level = "senior"
			t.tags = []string{"Microservices", "PostgreSQL", "Testing", "Security", "Docker"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "07_profile_me_api"):
			t.level = "junior"
			t.tags = []string{"REST", "Testing", "API Design"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "08_avatar"):
			t.level = "junior"
			t.tags = []string{"REST", "Testing", "Performance"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "09_notifications"):
			t.level = "middle"
			t.tags = []string{"Microservices", "REST", "Docker", "Testing"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "10_transactional_outbox"):
			t.level = "senior"
			t.tags = []string{"PostgreSQL", "Microservices", "Docker", "CI/CD", "Testing"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "11_rate_limit"):
			t.level = "senior"
			t.tags = []string{"Security", "Performance", "Redis", "Testing", "CI/CD"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "12_audit"):
			t.level = "middle"
			t.tags = []string{"Security", "PostgreSQL", "REST", "Testing"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "13_objects_crud"):
			t.level = "middle"
			t.tags = []string{"REST", "PostgreSQL", "Testing", "API Design"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "14_health"):
			t.level = "senior"
			t.tags = []string{"Docker", "CI/CD", "Performance", "PostgreSQL", "Testing", "Redis"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "15_admin"):
			t.level = "middle"
			t.tags = []string{"REST", "Security", "PostgreSQL", "Testing", "API Design"}
			t.languages = []string{"Go"}
		case strings.Contains(f, "16_pizza_api"):
			t.taskName = "pizza-api"
			t.level = "middle"
			t.tags = []string{"REST", "PostgreSQL", "Testing", "API Design"}
			t.languages = []string{"Go"}
		}
		tasks = append(tasks, t)
	}

	frontendFiles := readDir(frontendDir)
	for _, f := range frontendFiles {
		content := readFile(filepath.Join(frontendDir, f))
		title := extractTitle(content)
		description := extractDescription(content)

		t := task{file: f, taskName: taskNameFromFile(f), title: title, description: description, taskType: "frontend"}

		switch {
		case strings.Contains(f, "01_auth_pages"):
			t.level = "middle"
			t.tags = []string{"React", "Testing", "REST", "Security"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "02_password_recovery"):
			t.level = "junior"
			t.tags = []string{"React", "Testing", "REST"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "03_header"):
			t.level = "middle"
			t.tags = []string{"React", "Testing", "Performance"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "04_profile"):
			t.level = "junior"
			t.tags = []string{"React", "Testing", "REST"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "05_settings"):
			t.level = "middle"
			t.tags = []string{"React", "Testing", "Security"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "06_design"):
			t.level = "senior"
			t.tags = []string{"React", "Testing", "CI/CD", "Performance"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "07_api_client"):
			t.level = "senior"
			t.tags = []string{"React", "Testing", "Security", "REST"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "08_protected"):
			t.level = "middle"
			t.tags = []string{"React", "Testing", "Security"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "09_objects_table"):
			t.level = "middle"
			t.tags = []string{"React", "Testing", "REST", "API Design"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "10_object_editor"):
			t.level = "senior"
			t.tags = []string{"React", "Testing", "API Design", "REST"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "11_toasts"):
			t.level = "junior"
			t.tags = []string{"React", "Testing", "Performance"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "12_responsive"):
			t.level = "middle"
			t.tags = []string{"React", "Testing", "Performance"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "13_component"):
			t.level = "senior"
			t.tags = []string{"React", "Testing", "CI/CD", "Performance"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "14_theme"):
			t.level = "junior"
			t.tags = []string{"React", "Testing", "Performance"}
			t.languages = []string{"TypeScript"}
		case strings.Contains(f, "15_frontend_e2e"):
			t.level = "middle"
			t.tags = []string{"Testing", "CI/CD", "React", "REST"}
			t.languages = []string{"TypeScript"}
		}
		tasks = append(tasks, t)
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].file < tasks[j].file
	})
	ensureUniqueTaskNames(tasks)

	var sb strings.Builder
	sb.WriteString("-- +goose Up\n\n")

	// Collect unique tags and languages across all tasks for reference data.
	tagSet := make(map[string]struct{})
	langSet := make(map[string]struct{})
	for _, t := range tasks {
		for _, tag := range t.tags {
			tagSet[tag] = struct{}{}
		}
		for _, lang := range t.languages {
			langSet[lang] = struct{}{}
		}
	}
	tagNames := sortedKeys(tagSet)
	langNames := sortedKeys(langSet)

	sb.WriteString("INSERT INTO tags (id, name) VALUES\n")
	for i, name := range tagNames {
		tagID := fmt.Sprintf("c2000000-0000-0000-0000-%012d", i+1)
		sep := ","
		if i == len(tagNames)-1 {
			sep = ";"
		}
		sb.WriteString(fmt.Sprintf("  ('%s', %s)%s\n", tagID, escapeSQL(name), sep))
	}
	sb.WriteString("\n")

	sb.WriteString("INSERT INTO languages (id, name) VALUES\n")
	for i, name := range langNames {
		langID := fmt.Sprintf("c3000000-0000-0000-0000-%012d", i+1)
		sep := ","
		if i == len(langNames)-1 {
			sep = ";"
		}
		sb.WriteString(fmt.Sprintf("  ('%s', %s)%s\n", langID, escapeSQL(name), sep))
	}
	sb.WriteString("\n")

	for i, t := range tasks {
		taskID := fmt.Sprintf("c1000000-0000-0000-0000-%012d", i+1)

		sb.WriteString(fmt.Sprintf("INSERT INTO tasks (id, task_name, title, description, specification_md_text, task_type, level) VALUES\n"))
		sb.WriteString(fmt.Sprintf("  ('%s', %s, %s, %s, %s, '%s', '%s');\n\n",
			taskID,
			escapeSQL(t.taskName),
			escapeSQL(t.title),
			escapeSQL(t.description),
			escapeSQL(readFile(filepath.Join("test_data/spec_markdown", t.taskType, t.file))),
			t.taskType,
			t.level,
		))

		for _, tagName := range t.tags {
			sb.WriteString(fmt.Sprintf("INSERT INTO task_tags (task_id, tag_id) SELECT '%s', id FROM tags WHERE name = '%s';\n", taskID, escapeSQLRaw(tagName)))
		}
		sb.WriteString("\n")

		for _, langName := range t.languages {
			sb.WriteString(fmt.Sprintf("INSERT INTO task_languages (task_id, language_id) SELECT '%s', id FROM languages WHERE name = '%s';\n", taskID, escapeSQLRaw(langName)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("-- +goose Down\n\n")
	sb.WriteString("DELETE FROM task_languages;\n")
	sb.WriteString("DELETE FROM task_tags;\n")
	sb.WriteString("DELETE FROM tasks;\n")
	sb.WriteString("DELETE FROM languages;\n")
	sb.WriteString("DELETE FROM tags;\n")

	os.WriteFile("migrations/20260622000005_seed_test_tasks.sql", []byte(sb.String()), 0644)
	fmt.Printf("Generated migration with %d tasks\n", len(tasks))
}

func readDir(dir string) []string {
	entries, _ := os.ReadDir(dir)
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files
}

func readFile(path string) string {
	data, _ := os.ReadFile(path)
	return string(data)
}

func extractTitle(content string) string {
	lines := strings.SplitN(content, "\n", 2)
	title := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(lines[0], "# ТЗ: "), "# "))
	if title != "Техническое задание" {
		return title
	}
	if projectName := extractSectionFirstText(content, "## Название проекта"); projectName != "" {
		return projectName
	}
	return title
}

func extractDescription(content string) string {
	if context := extractSectionFirstText(content, "## Контекст"); context != "" {
		return context
	}
	return extractSectionFirstText(content, "## Цель")
}

func extractSectionFirstText(content, section string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) == section {
			for j := i + 1; j < len(lines); j++ {
				l := strings.TrimSpace(lines[j])
				if l != "" {
					return l
				}
			}
		}
	}
	return ""
}

func taskNameFromFile(file string) string {
	name := strings.TrimSuffix(file, filepath.Ext(file))
	parts := strings.SplitN(name, "_", 2)
	if len(parts) == 2 {
		name = parts[1]
	}
	return strings.ReplaceAll(name, "_", "-")
}

func ensureUniqueTaskNames(tasks []task) {
	seen := make(map[string]bool, len(tasks))
	for i := range tasks {
		if !seen[tasks[i].taskName] {
			seen[tasks[i].taskName] = true
			continue
		}
		tasks[i].taskName = tasks[i].taskType + "-" + tasks[i].taskName
		seen[tasks[i].taskName] = true
	}
}

func escapeSQL(s string) string {
	s = strings.ReplaceAll(s, "'", "''")
	return "'" + s + "'"
}

func escapeSQLRaw(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
