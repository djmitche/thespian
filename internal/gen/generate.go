package gen

import (
	"io/ioutil"

	"github.com/iancoleman/strcase"
	"golang.org/x/tools/go/packages"
	"gopkg.in/yaml.v2"
)

type ThespianYml struct {
	Actors    map[string]ActorYml   `yaml:"actors"`
	Mailboxes map[string]MailboxYml `yaml:"mailboxes"`
}

type ActorYml struct {
	Test      bool                       `yaml:"test"`
	Mailboxes map[string]ActorMailboxYml `yaml:"mailboxes"`
}

type ActorMailboxYml struct {
	Kind         string `yaml:"kind"`
	MessageType  string `yaml:"message-type"`
	RequestType  string `yaml:"request-type"`
	ResponseType string `yaml:"response-type"`
	Import       string `yaml:"import"`
	Type         string `yaml:"type"`
}

type MailboxYml struct {
	Test         bool   `yaml:"test"`
	Kind         string `yaml:"kind"`
	MessageType  string `yaml:"message-type"`
	RequestType  string `yaml:"request-type"`
	ResponseType string `yaml:"response-type"`
}

func Generate() {
	// get the name of the current package
	cfg := &packages.Config{Mode: packages.NeedName}
	pkgs, err := packages.Load(cfg)
	if err != nil {
		bail("Could not determine current package")
	}
	thisPackageName := pkgs[0].Name

	rawYml, err := ioutil.ReadFile("thespian.yml")
	if err != nil {
		bail("Could not load thespian.yml: %s", err)
	}

	var yml ThespianYml
	err = yaml.Unmarshal(rawYml, &yml)
	if err != nil {
		bail("Could not parse thespian.yml: %s", err)
	}

	filename := func(typeName string, test bool) string {
		if test {
			return strcase.ToSnake(typeName) + "_thespian_gen_test.go"
		}
		return strcase.ToSnake(typeName) + "_thespian_gen.go"
	}

	for actorName, actor := range yml.Actors {
		out := newFormatter(filename(actorName, actor.Test))
		actGen := NewActorGenerator(thisPackageName, actorName, actor)
		actGen.GenerateGo(out)
		err = out.write()
		if err != nil {
			bail("Error: %s", err)
		}
	}

	for mboxName, mbox := range yml.Mailboxes {
		out := newFormatter(filename(mboxName, mbox.Test))
		mbGen := NewMailboxGeneratorForMailbox(thisPackageName, mboxName, mbox)
		mbGen.GenerateGo(out)
		err = out.write()
		if err != nil {
			bail("Error: %s", err)
		}
	}
}
