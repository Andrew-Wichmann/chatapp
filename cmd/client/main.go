package main

import (
	"fmt"
	"strings"

	"github.com/Andrew-Wichmann/chatapp/pkg/client"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type serverResponse string

type errMsg struct{ error }

type App struct {
	err          error
	messageInput textarea.Model
	chatHistory  viewport.Model
	rpcClient    client.ChatAppClient
	history      []string
}

func newApp() App {
	ta := textarea.New()

	ta.Placeholder = "Start chatting"
	ta.Prompt = "┃ "
	ta.ShowLineNumbers = false
	ta.SetHeight(1)
	ta.Focus()
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(30, 5)

	vp.SetContent("Welcome to the chatroom")

	return App{
		messageInput: ta,
		chatHistory:  vp,
	}
}

func (app App) Init() tea.Cmd {
	return textarea.Blink
}

func (app App) View() string {
	if app.err != nil {
		return fmt.Sprintf("something went wrong: %s", app.err)
	}
	return fmt.Sprintf("%s\n%s\nPress CTRL+C to exit", app.chatHistory.View(), app.messageInput.View())
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		messageCmd     tea.Cmd
		chatHistoryCmd tea.Cmd
	)

	app.messageInput, messageCmd = app.messageInput.Update(msg)
	app.chatHistory, chatHistoryCmd = app.chatHistory.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return app, tea.Quit
		case tea.KeyEnter:
			cmd := sendMessage(app)
			app.messageInput.Reset()
			return app, cmd
		}
	case errMsg:
		app.err = msg
	case serverResponse:
		app.history = append(app.history, string(msg))
		app.chatHistory.SetContent(strings.Join(app.history, "\n"))
	}
	return app, tea.Batch(messageCmd, chatHistoryCmd)
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
		message, err := app.rpcClient.SendMessageRPC(app.messageInput.Value())
		if err != nil {
			return errMsg{err}
		}
		return serverResponse(message)
	}
}
