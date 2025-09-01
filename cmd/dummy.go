package main

import tea "github.com/charmbracelet/bubbletea"

type Water struct{}

func (m Water) Init() tea.Cmd                           { return nil }
func (m Water) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m Water) View() string                            { return "Water" }
