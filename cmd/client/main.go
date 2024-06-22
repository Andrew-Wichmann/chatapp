package main

import tea "github.com/charmbracelet/bubbletea"


type App struct{}

func (app App) Init() tea.Cmd {
    return nil
}

func (app App) View() string {
    return "Press CTRL+C to exit"
}

func (app App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type){
    case tea.KeyMsg:
        switch msg.Type{
        case tea.KeyCtrlC:
            return app, tea.Quit
        }
    }
    return app, nil
}

func main() {
    app := App{}
    prog := tea.NewProgram(app)

    prog.Run()
}
