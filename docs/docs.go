// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/admin/message": {
            "put": {
                "security": [
                    {
                        "AdminUserAuth": []
                    }
                ],
                "description": "Updates a message",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "Admin"
                ],
                "operationId": "UpdateMessage",
                "parameters": [
                    {
                        "description": "body json",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.Message"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Message"
                        }
                    }
                }
            },
            "post": {
                "security": [
                    {
                        "AdminUserAuth": []
                    }
                ],
                "description": "Creates a message",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "Admin"
                ],
                "operationId": "CreateMessage",
                "parameters": [
                    {
                        "description": "body json",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.Message"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Message"
                        }
                    }
                }
            }
        },
        "/admin/message/{id}": {
            "get": {
                "security": [
                    {
                        "AdminUserAuth": []
                    }
                ],
                "description": "Retrieves a message by id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Admin"
                ],
                "operationId": "GetMessage",
                "parameters": [
                    {
                        "type": "string",
                        "description": "id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Message"
                        }
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "AdminUserAuth": []
                    }
                ],
                "description": "Deletes a message with id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Admin"
                ],
                "operationId": "DeleteMessage",
                "parameters": [
                    {
                        "type": "string",
                        "description": "id",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/admin/messages": {
            "get": {
                "security": [
                    {
                        "AdminUserAuth": []
                    }
                ],
                "description": "Gets all messages",
                "tags": [
                    "Admin"
                ],
                "operationId": "GetMessages",
                "parameters": [
                    {
                        "type": "string",
                        "description": "uin - filter by uin",
                        "name": "uin",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "email - filter by email",
                        "name": "email",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "phone - filter by phone",
                        "name": "phone",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "topic - filter by topic",
                        "name": "topic",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "offset",
                        "name": "offset",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "limit - limit the result",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "order - Possible values: asc, desc. Default: desc",
                        "name": "order",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/model.Message"
                            }
                        }
                    }
                }
            }
        },
        "/admin/topic": {
            "put": {
                "security": [
                    {
                        "AdminUserAuth": []
                    }
                ],
                "description": "Updated the topic.",
                "tags": [
                    "Admin"
                ],
                "operationId": "UpdateTopic",
                "parameters": [
                    {
                        "description": "body json",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/Topic"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/Topic"
                        }
                    }
                }
            }
        },
        "/admin/topics": {
            "get": {
                "security": [
                    {
                        "AdminUserAuth": []
                    }
                ],
                "description": "Gets all topics",
                "tags": [
                    "Admin"
                ],
                "operationId": "AdminGetTopics",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/Topic"
                            }
                        }
                    }
                }
            }
        },
        "/int/message": {
            "post": {
                "security": [
                    {
                        "InternalAuth": []
                    }
                ],
                "description": "Sends a message to a user, list of users or a topic",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Internal"
                ],
                "operationId": "InternalSendMessage",
                "parameters": [
                    {
                        "description": "body json",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.Message"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Message"
                        }
                    }
                }
            }
        },
        "/message": {
            "post": {
                "security": [
                    {
                        "AdminUserAuth": []
                    }
                ],
                "description": "Sends a message to a user, list of users or a topic",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "Client"
                ],
                "operationId": "SendMessage",
                "parameters": [
                    {
                        "description": "body json",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.Message"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.Message"
                        }
                    }
                }
            }
        },
        "/messages": {
            "get": {
                "security": [
                    {
                        "RokwireAuth": []
                    }
                ],
                "description": "Gets all messages to the authenticated user.",
                "tags": [
                    "Client"
                ],
                "operationId": "GetUserMessages",
                "parameters": [
                    {
                        "type": "string",
                        "description": "offset",
                        "name": "offset",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "limit - limit the result",
                        "name": "limit",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "order - Possible values: asc, desc. Default: desc",
                        "name": "order",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/model.Message"
                            }
                        }
                    }
                }
            }
        },
        "/token": {
            "post": {
                "security": [
                    {
                        "RokwireAuth": []
                    }
                ],
                "description": "Stores a firebase token and maps it to a idToken if presents",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "Client"
                ],
                "operationId": "Token",
                "parameters": [
                    {
                        "description": "body json",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/tokenBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/topic/{topic}/messages": {
            "get": {
                "security": [
                    {
                        "RokwireAuth": []
                    }
                ],
                "description": "Gets all messages for topic",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Client"
                ],
                "operationId": "GetTopicMessages",
                "parameters": [
                    {
                        "type": "string",
                        "description": "topic",
                        "name": "topic",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/model.Message"
                            }
                        }
                    }
                }
            }
        },
        "/topic/{topic}/subscribe": {
            "post": {
                "security": [
                    {
                        "RokwireAuth": []
                    }
                ],
                "description": "Subscribes the current user to a topic",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "Client"
                ],
                "operationId": "Subscribe",
                "parameters": [
                    {
                        "type": "string",
                        "description": "topic",
                        "name": "topic",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "body json",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/tokenBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/topic/{topic}/unsubscribe": {
            "post": {
                "security": [
                    {
                        "RokwireAuth": []
                    }
                ],
                "description": "Unsubscribes the current user to a topic",
                "tags": [
                    "Client"
                ],
                "operationId": "Unsubscribe",
                "parameters": [
                    {
                        "type": "string",
                        "description": "topic",
                        "name": "topic",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "body json",
                        "name": "data",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/tokenBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        },
        "/topics": {
            "get": {
                "security": [
                    {
                        "RokwireAuth": []
                    }
                ],
                "description": "Gets all topics",
                "tags": [
                    "Client"
                ],
                "operationId": "GetTopics",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/Topic"
                            }
                        }
                    }
                }
            }
        },
        "/version": {
            "get": {
                "security": [
                    {
                        "RokwireAuth": []
                    }
                ],
                "description": "Gives the service version.",
                "produces": [
                    "text/plain"
                ],
                "tags": [
                    "Client"
                ],
                "operationId": "Version",
                "responses": {
                    "200": {
                        "description": ""
                    }
                }
            }
        }
    },
    "definitions": {
        "Recipient": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "uin": {
                    "type": "string"
                }
            }
        },
        "Topic": {
            "type": "object",
            "properties": {
                "date_created": {
                    "type": "string"
                },
                "date_updated": {
                    "type": "string"
                },
                "description": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "User": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "phone": {
                    "type": "string"
                },
                "uiucedu_is_member_of": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "uiucedu_uin": {
                    "type": "string"
                }
            }
        },
        "model.Message": {
            "type": "object",
            "properties": {
                "body": {
                    "type": "string"
                },
                "date_created": {
                    "type": "string"
                },
                "date_sent": {
                    "type": "string"
                },
                "date_updated": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "recipients": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/Recipient"
                    }
                },
                "sender": {
                    "$ref": "#/definitions/model.Sender"
                },
                "sent": {
                    "type": "boolean"
                },
                "subject": {
                    "type": "string"
                },
                "topic": {
                    "type": "string"
                }
            }
        },
        "model.Sender": {
            "type": "object",
            "properties": {
                "type": {
                    "description": "user or system",
                    "type": "string"
                },
                "user": {
                    "$ref": "#/definitions/User"
                }
            }
        },
        "tokenBody": {
            "type": "object",
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "AdminUserAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header (add Bearer prefix to the Authorization value)"
        },
        "InternalAuth": {
            "type": "apiKey",
            "name": "INTERNAL-API-KEY",
            "in": "header"
        },
        "RokwireAuth": {
            "type": "apiKey",
            "name": "ROKWIRE-API-KEY",
            "in": "header"
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "0.1.0",
	Host:        "localhost",
	BasePath:    "/notifications/api",
	Schemes:     []string{"https"},
	Title:       "Rokwire Notifications Building Block API",
	Description: "Rokwire Notifications Building Block API Documentation.",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
