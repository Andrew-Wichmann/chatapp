package main

import (
	"fmt"

	"github.com/Andrew-Wichmann/chatapp/pkg/client"
	tea "github.com/charmbracelet/bubbletea"
)

type serverResponse string

type errMsg struct{ error }

type App struct {
	err       error
	message   string
	rpcClient client.ChatAppClient
}

func (app App) Init() tea.Cmd {
	closure := func() tea.Msg {
		return checkServer(app)
	}
	return closure
}

func (app App) View() string {
	if app.message != "" {
		return fmt.Sprintf("Done! Got: %s. Press CTRL+C to exit", app.message)
	}
	if app.err != nil {
		return fmt.Sprintf("something went wrong: %s", app.err)
	}
	return "Processing. Press CTRL+C to exit"
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return app, tea.Quit
		}
	case errMsg:
		app.err = msg
	case serverResponse:
		app.message = string(msg)
	}
	return app, nil
}

func main() {
	app := App{}
	prog := tea.NewProgram(app)

	prog.Run()
}

func checkServer(app App) tea.Msg {
	message, err := app.rpcClient.SendMessageRPC("foobar")

	if err != nil {
		return errMsg{err}
	}

	return serverResponse(message)
}
