package gen

import "fmt"

type MailboxGenerator interface {
	GenerateGo(*formatter)
	ActorBuilderStructDecl() string
	ActorRxStructDecl() string
	ActorRxInitializer() string
	ActorTxStructDecl() string
	ActorTxInitializer() string
	ActorTxStructMethod() string
	ActorSpawnSetupClause() string
	ActorLoopCase() string
	ActorCleanupClause() string
}

func NewMailboxGeneratorForActor(thisPackage, actorName, fieldName string, imports *importTracker, yml ActorMailboxYml) MailboxGenerator {
	mboxPkg, mboxType := SplitPackage(yml.Type)
	mboxTypeQual := ""
	if mboxPkg != "" {
		shortName := imports.add(mboxPkg)
		mboxTypeQual = fmt.Sprintf("%s.", shortName)
	}

	switch yml.Kind {
	case "simple":
		msgPkg, msgType := SplitPackage(yml.MessageType)
		if msgPkg != "" {
			shortName := imports.add(msgPkg)
			msgType = fmt.Sprintf("%s.%s", shortName, msgType)
		}

		return &SimpleMailboxGenerator{
			ThisPackage:  thisPackage,
			MessageType:  msgType,
			MboxTypeBase: mboxType,
			MboxTypeQual: mboxTypeQual,
			ActorName:    actorName,
			FieldName:    fieldName,
		}
	case "ticker":
		return &TickerMailboxGenerator{
			ThisPackage:  thisPackage,
			MboxTypeBase: mboxType,
			MboxTypeQual: mboxTypeQual,
			ActorName:    actorName,
			FieldName:    fieldName,
		}
	default:
		bail("unknown mailbox kind %s", yml.Kind)
		return nil
	}
}

func NewMailboxGeneratorForMailbox(thisPackage, typeName string, yml MailboxYml) MailboxGenerator {
	switch yml.Kind {
	case "simple":
		imports := newImportTracker()

		msgPkg, msgType := SplitPackage(yml.MessageType)
		if msgPkg != "" {
			shortName := imports.add(msgPkg)
			msgType = fmt.Sprintf("%s.%s", shortName, msgType)
		}

		return &SimpleMailboxGenerator{
			ThisPackage:  thisPackage,
			MessageType:  msgType,
			MboxTypeBase: typeName,
			Imports:      imports.get(),
		}
	case "ticker":
		return &TickerMailboxGenerator{
			ThisPackage:  thisPackage,
			MboxTypeBase: typeName,
		}
	default:
		bail("unknown template kind %s", yml.Kind)
		return nil
	}
}
