package gen

import (
	"io/ioutil"

	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v2"
)

type ThespianYml struct {
	Package   string                `yaml:"package"`
	Actors    map[string]ActorYml   `yaml:"actors"`
	Mailboxes map[string]MailboxYml `yaml:"mailboxes"`
}

type ActorYml struct {
	Mailboxes map[string]ActorMailboxYml `yaml:"mailboxes"`
}

type ActorMailboxYml struct {
	Kind        string `yaml:"kind"`
	MessageType string `yaml:"message-type"`
	Package     string `yaml:"package"`
	Type        string `yaml:"type"`
}

type MailboxYml struct {
	Kind        string `yaml:"kind"`
	MessageType string `yaml:"message-type"`
}

func Generate() {
	rawYml, err := ioutil.ReadFile("thespian.yml")
	if err != nil {
		bail("Could not load thespian.yml: %s", err)
	}

	var yml ThespianYml
	err = yaml.Unmarshal(rawYml, &yml)
	if err != nil {
		bail("Could not parse thespian.yml: %s", err)
	}

	for name, actor := range yml.Actors {
		out := newFormatter(strcase.ToSnake(name) + "_thespian_gen.go")
		actor.Generate(yml.Package, name, out)
		err = out.write()
		if err != nil {
			bail("Error: %s", err)
		}
	}

	for name, mbox := range yml.Mailboxes {
		out := newFormatter(strcase.ToSnake(name) + "_thespian_gen.go")
		mbox.Generate(yml.Package, name, out)
		err = out.write()
		if err != nil {
			bail("Error: %s", err)
		}
	}
}
