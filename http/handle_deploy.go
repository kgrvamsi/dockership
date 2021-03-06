package http

import (
	"encoding/json"
	"time"

	"github.com/mcuadros/dockership/core"

	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

func (s *server) HandleDeploy(msg Message, session sockjs.Session) {
	force := true
	project, ok := msg.Request["project"]
	if !ok {
		core.Error("Missing project", "request", "deploy")
		return
	}

	environment, ok := msg.Request["environment"]
	if !ok {
		core.Error("Missing environment", "request", "deploy")
		return
	}

	writer := NewSockJSWriter(s.sockjs, "deploy")
	now := time.Now()
	writer.SetFormater(func(raw []byte) []byte {
		str, _ := json.Marshal(map[string]string{
			"environment": environment,
			"project":     project,
			"date":        now.String(),
			"log":         string(raw),
		})

		return str
	})

	if p, ok := s.config.Projects[project]; ok {
		core.Info("Starting deploy", "project", p, "environment", environment, "force", force)

		go func(session sockjs.Session) {
			time.Sleep(50 * time.Millisecond)
			s.EmitProjects(session)
		}(session)

		err := p.Deploy(environment, writer, force)
		if len(err) != 0 {
			for _, e := range err {
				core.Critical(e.Error(), "project", project)
			}
		} else {
			core.Info("Deploy success", "project", p, "environment", environment)
		}
	} else {
		core.Error("Project not found", "project", p)
	}

	s.EmitProjects(session)
}
