package main

import (
	"log"
	"log-service/data"
)

type RPCServer struct {
	Models *data.Models
}

type RPCPayload struct {
	Name string
	Data string
}

func (s *RPCServer) LogInfo(payload *RPCPayload, resp *string) error {
	event := data.LogEntry{
		Name: payload.Name,
		Data: payload.Data,
	}

	err := s.Models.LogEntry.Insert(event)
	if err != nil {
		log.Println("Error inserting log entry:", err)
		return err
	}

	*resp = "Log entry created"
	return nil
}
