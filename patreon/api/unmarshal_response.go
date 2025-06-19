package api

import (
	"encoding/json"
	"fmt"
	"io"
)

func UnmarshalEntity(entityData json.RawMessage) (any, error) {
	var unmarshalStruct = struct {
		Type string `json:"type"`
	}{}
	err := json.Unmarshal(entityData, &unmarshalStruct)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal partial entity: %w", err)
	}

	var target any
	switch unmarshalStruct.Type {
	case "media":
		t := ResponseMedia{}
		err = json.Unmarshal(entityData, &t)
		target = t
		break
	case "member":
		t := ResponseMember{}
		err = json.Unmarshal(entityData, &t)
		target = t
		break
	case "user":
		t := ResponseUser{}
		err = json.Unmarshal(entityData, &t)
		target = t
		break
	case "post":
		t := ResponsePost{}
		err = json.Unmarshal(entityData, &t)
		target = t
		break
	case "reward":
		t := ResponseReward{}
		err = json.Unmarshal(entityData, &t)
		target = t
		break
	case "campaign":
		t := ResponseCampaign{}
		err = json.Unmarshal(entityData, &t)
		target = t
		break
	default:
		return nil, fmt.Errorf("unknown entity type: %s", unmarshalStruct.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal entity: %w", err)
	}

	return target, nil
}

func UnmarshalResponse[TData any](reader io.Reader, target *Response[TData]) error {
	unmarshalStruct := struct {
		Data     TData             `json:"data,omitempty"`
		Included []json.RawMessage `json:"included,omitempty"`
		Meta     ResponseMeta      `json:"meta,omitempty"`
		Links    ResponseLinks     `json:"links,omitempty"`
	}{}

	err := json.NewDecoder(reader).Decode(&unmarshalStruct)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	target.Data = unmarshalStruct.Data
	target.Meta = unmarshalStruct.Meta
	target.Links = unmarshalStruct.Links
	target.Included = make([]any, 0)

	for _, include := range unmarshalStruct.Included {
		entity, err := UnmarshalEntity(include)
		if err != nil {
			return fmt.Errorf("failed to unmarshal entity: %w", err)
		}
		target.Included = append(target.Included, entity)
	}

	return nil
}
