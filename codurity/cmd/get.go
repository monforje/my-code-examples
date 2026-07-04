package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"codurity/internal/auth"
	"codurity/internal/build"
	"codurity/internal/tasks"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:           "get <task_name>",
	Short:         "Get and clone a task repository",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		taskName := args[0]

		a, err := auth.Load()
		if err != nil {
			color.New(color.FgRed).Fprintf(out, "Ошибка: не авторизован.\n")
			color.New(color.FgYellow).Fprintf(out, "Выполните %s для входа.\n", color.New(color.Bold).Sprint("codurity login"))
			return nil
		}

		if a.IsExpired() {
			color.New(color.FgYellow).Fprintln(out, "Токен истёк, обновляем...")
			authClient := auth.NewClient(build.AuthAPIURL)
			ref, err := authClient.Refresh(cmd.Context(), a.RefreshToken)
			if err != nil {
				color.New(color.FgRed).Fprintf(out, "Ошибка обновления токена: %s\n", err)
				color.New(color.FgYellow).Fprintf(out, "Выполните %s для повторного входа.\n", color.New(color.Bold).Sprint("codurity login"))
				return nil
			}
			a.AccessToken = ref.AccessToken
			a.RefreshToken = ref.RefreshToken
			if err := a.Save(); err != nil {
				return fmt.Errorf("save refreshed auth: %w", err)
			}
		}

		tasksClient := tasks.NewClient(build.TasksAPIURL)

		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Suffix = fmt.Sprintf(" Получаем задачу %q...", taskName)
		s.Start()

		result, err := tasksClient.CreateGitTask(cmd.Context(), taskName, a.AccessToken)

		s.Stop()
		fmt.Fprintln(out)

		if err != nil {
			return printTaskError(out, err)
		}

		color.New(color.FgGreen).Fprintf(out, "✔ Задача:  %s\n", result.TaskName)
		fmt.Fprintf(out, "  Репо:   %s\n", result.Repo)
		fmt.Fprintln(out)

		if !askYes(out, fmt.Sprintf("Склонировать задачу в папку %s?", color.New(color.Bold, color.FgCyan).Sprint(result.Repo))) {
			color.New(color.FgYellow).Fprintln(out, "Вызовите эту команду там где хотите и согласитесь.")
			return nil
		}

		fmt.Fprintln(out)
		s = spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Suffix = " Клонируем..."
		s.Start()

		gitCmd := exec.CommandContext(cmd.Context(), "git", "clone", result.CloneURL)
		gitCmd.Stdout = out
		gitCmd.Stderr = out
		cloneErr := gitCmd.Run()

		s.Stop()
		fmt.Fprintln(out)

		if cloneErr != nil {
			color.New(color.FgRed).Fprintf(out, "Ошибка клонирования: %s\n", cloneErr)
			return nil
		}

		color.New(color.FgGreen).Fprintf(out, "✔ Репозиторий %s успешно клонирован.\n", result.Repo)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}

func printTaskError(out io.Writer, err error) error {
	var ae *tasks.APIError
	if errors.As(err, &ae) {
		switch ae.StatusCode {
		case 400:
			color.New(color.FgRed).Fprintf(out, "Ошибка валидации: %s\n", ae.Message)
		case 401:
			color.New(color.FgRed).Fprintln(out, "Ошибка: не авторизован или токен недействителен.")
			color.New(color.FgYellow).Fprintf(out, "Выполните %s для входа.\n", color.New(color.Bold).Sprint("codurity login"))
		case 404:
			color.New(color.FgRed).Fprintf(out, "Задача %q не найдена.\n", ae.Message)
		case 409:
			color.New(color.FgYellow).Fprintf(out, "Конфликт: %s\n", ae.Message)
		default:
			color.New(color.FgRed).Fprintf(out, "Ошибка сервера (%d): %s\n", ae.StatusCode, ae.Message)
		}
		return nil
	}
	color.New(color.FgRed).Fprintf(out, "Ошибка: %s\n", err)
	return nil
}

func askYes(out io.Writer, prompt string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		color.New(color.FgCyan).Fprintf(out, "%s ", prompt)
		color.New(color.Faint).Fprint(out, "[Да/Нет]: ")

		answer, err := reader.ReadString('\n')
		if err != nil {
			return false
		}
		answer = strings.TrimSpace(strings.ToLower(answer))
		switch answer {
		case "да", "yes", "y", "д", "d":
			return true
		case "нет", "no", "n", "н":
			return false
		}
	}
}
