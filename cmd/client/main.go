package main

import (
	"fmt"

	"github.com/Andrew-Wichmann/chatapp/pkg/client"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type serverResponse string

type errMsg struct{ error }

type App struct {
	err       error
	message   string
	ta        textarea.Model
	rpcClient client.ChatAppClient
}

func newApp() App {
	ta := textarea.New()

	ta.Placeholder = "Start chatting"
	ta.Prompt = "┃ "
	ta.ShowLineNumbers = false
	ta.Focus()

	return App{
		ta: ta,
	}
}

func (app App) Init() tea.Cmd {
	return textarea.Blink
}

func (app App) View() string {
	if app.message != "" {
		return fmt.Sprintf("Done! Got: %s. Press CTRL+C to exit", app.message)
	}
	if app.err != nil {
		return fmt.Sprintf("something went wrong: %s", app.err)
	}
	if app.ta.Placeholder != "" {
		return fmt.Sprintf("%s\nPress CTRL+C to exit", app.ta.View())
	}
	return "Initializing"
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
	)

	app.ta, tiCmd = app.ta.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return app, tea.Quit
		case tea.KeyEnter:
			cmd := sendMessage(app)
			app.ta.Reset()
			return app, cmd
		}
	case errMsg:
		app.err = msg
	case serverResponse:
		app.message = string(msg)
	}
	return app, tiCmd
}

func main() {
	prog := tea.NewProgram(newApp())

	prog.Run()
}

func checkServer(app App) tea.Msg {
	message, err := app.rpcClient.SendMessageRPC("foobar")

	if err != nil {
		return errMsg{err}
	}

	return serverResponse(message)
}

func sendMessage(app App) tea.Cmd {
	return func() tea.Msg {
		message, err := app.rpcClient.SendMessageRPC(app.ta.Value())
		if err != nil {
			return errMsg{err}
		}
		return serverResponse(message)
	}
}
