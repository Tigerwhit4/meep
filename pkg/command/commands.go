package command

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/StreamMeBots/meep/pkg/buckets"
	"github.com/StreamMeBots/meep/pkg/db"
	"github.com/StreamMeBots/pkg/commands"

	"github.com/boltdb/bolt"
)

// Errors
var ErrCommandNotFound = errors.New("Command not found")

// Command represents a command response template.
type Command struct {
	Name     string `json:"name"`
	Template string `json:"template"`
}

// Validate validates the Command
func (c *Command) Validate() error {
	if len(c.Name) == 0 || len(c.Name) > 100 {
		return fmt.Errorf("Command Name should be between 1 and 100 characters")
	}
	if len(c.Template) == 0 || len(c.Template) > 500 {
		return fmt.Errorf("Command Template should be between 1 and 500 characters")
	}

	if _, err := template.New("foo").Parse(c.Template); err != nil {
		return fmt.Errorf("Error parsing Template: %v", err)
	}

	return nil
}

// Save saves the command
func (c *Command) Save(userBucket []byte) error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		b, err := json.Marshal(c)
		if err != nil {
			return err
		}

		bkt, err := buckets.UserCommands(userBucket, tx)
		if err != nil {
			return err
		}

		return bkt.Put([]byte(c.Name), b)
	})

	if err != nil {
		log.Println("msg='error-saving-command', error='%v', userBucket='%s'", err, string(userBucket))
		return err
	}

	return nil
}

// Get gets a single command
func Get(userBucket []byte, name string) (*Command, error) {
	var cmd *Command
	err := db.DB.Update(func(tx *bolt.Tx) error {
		bkt, err := buckets.UserCommands(userBucket, tx)
		if err != nil {
			return err
		}

		b := bkt.Get([]byte(name))
		if b == nil {
			return nil
		}

		if err := json.Unmarshal(b, &cmd); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("msg='error-reading-command', error='%v', userBucket='%s'\n", err, string(userBucket))
		return nil, err
	}

	if cmd == nil {
		return nil, ErrCommandNotFound
	}

	return cmd, nil
}

// GetAll gets all of a user's commands
func GetAll(userBucket []byte) ([]*Command, error) {
	cmds := []*Command{}

	err := db.DB.Update(func(tx *bolt.Tx) error {
		bkt, err := buckets.UserCommands(userBucket, tx)
		if err != nil {
			return err
		}

		bkt.ForEach(func(k, v []byte) error {
			cmd := &Command{}
			if err := json.Unmarshal(v, &cmd); err != nil {
				log.Println("msg='json-unmarshal-error', key='%v' value='%v' error='%v'", string(k), string(v), err)
				return nil
			}

			cmds = append(cmds, cmd)

			return nil
		})

		return nil
	})

	if err != nil {
		log.Println("msg='error-reading-command', error='%v', userBucket='%s'", err, string(userBucket))
		return nil, err
	}

	return cmds, nil
}

// Delete deletes a command from a user's bucket
func Delete(userBucket []byte, name string) error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		bkt, err := buckets.UserCommands(userBucket, tx)
		if err != nil {
			return err
		}

		return bkt.Delete([]byte(name))
	})

	if err != nil {
		log.Println("msg='error-saving-command', error='%v', userBucket='%s'", err, string(userBucket))
		return err
	}

	return nil
}

// Say checks if the message is a command and if it is provides an answer to the command
func Say(userBucket []byte, cmd *commands.Command) string {
	s := strings.TrimSpace(cmd.Get("message"))
	c, err := Get(userBucket, s)
	if err != nil {
		return ""
	}

	return parseTemplate(c.Template, cmd.Args)
}

func parseTemplate(tmpl string, d interface{}) string {
	t, err := template.New("msg").Parse(tmpl)
	if err != nil {
		log.Println("msg='error parsing template', template='%s', error='%v'", tmpl, err)
		return ""
	}

	buf := &bytes.Buffer{}
	if err := t.Execute(buf, d); err != nil {
		log.Println("msg='error executing template', template='%s', data='%+v', error='%v'", tmpl, d, err)
		return ""
	}

	return buf.String()
}
