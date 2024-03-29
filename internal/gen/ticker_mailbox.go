package gen

import "strings"

type TickerMailboxGenerator struct {
	ThisPackage  string
	MboxTypeBase string

	// actor generation only
	MboxTypeQual string
	ActorName    string
	FieldName    string
}

func (g *TickerMailboxGenerator) GenerateGo(out *formatter) {
	out.executeTemplate(getTemplate(
		"ticker_mailbox_go_file", `
// code generated by thespian; DO NOT EDIT

package {{.ThisPackage}}

import (
	"time"
	"github.com/benbjohnson/clock"
	"github.com/djmitche/thespian"
)

// {{.MboxTypeBase}}Rx contains a ticker that the actor implementation can control
type {{.MboxTypeBase}}Rx struct {
	// ticker is the ticker this mailbox responds to, or nil if it is disabled
	ticker *clock.Ticker
	clock clock.Clock
}

func New{{.MboxTypeBase}}Rx(rt *thespian.Runtime) {{.MboxTypeBase}}Rx {
	return {{.MboxTypeBase}}Rx{
		ticker: nil,
		clock: rt.Clock,
	}
}

// Chan gets a channel for this ticker, or nil if the ticker is not started.
func (rx *{{.MboxTypeBase}}Rx) Chan() <-chan time.Time {
	if rx.ticker == nil {
		return nil
	}
	return rx.ticker.C
}

// Reset stops a ticker and resets its period to the specified duration.
func (rx *{{.MboxTypeBase}}Rx) Reset(d time.Duration) {
	if rx.ticker == nil {
		rx.ticker = rx.clock.Ticker(d)
	} else {
		rx.ticker.Reset(d)
	}
}

// Stop stops this ticker.  This is called automatically on agent stop.
func (rx *{{.MboxTypeBase}}Rx) Stop() {
	if rx.ticker != nil {
		rx.ticker.Stop()
	}
}`), g)
}

func (g *TickerMailboxGenerator) ActorBuilderStructDecl() string {
	return ""
}

func (g *TickerMailboxGenerator) ActorRxStructDecl() string {
	return renderTemplate(
		"ticker_actor_rx_struct_decl",
		`{{.FieldName}} {{.MboxTypeQual}}{{.MboxTypeBase}}Rx`,
		g)
}

func (g *TickerMailboxGenerator) ActorRxInitializer() string {
	return renderTemplate(
		"ticker_actor_rx_initializer",
		`{{.FieldName}}: {{.MboxTypeQual}}New{{.MboxTypeBase}}Rx(rt),`,
		g)
}

func (g *TickerMailboxGenerator) ActorTxStructDecl() string {
	return ""
}

func (g *TickerMailboxGenerator) ActorTxInitializer() string {
	return ""
}

func (g *TickerMailboxGenerator) ActorTxStructMethod() string {
	return ""
}

func (g *TickerMailboxGenerator) ActorSpawnSetupClause() string {
	return ""
}

func (g *TickerMailboxGenerator) ActorLoopCase() string {
	return renderTemplate(
		"ticker_actor_loop_case", strings.TrimSpace(`
			case t := <-rx.{{.FieldName}}.Chan():
				a.handle{{public .FieldName}}(t)
		`), g)
}

func (g *TickerMailboxGenerator) ActorCleanupClause() string {
	return renderTemplate(
		"ticker_actor_cleanup_clause", strings.TrimSpace(`
			rx.{{.FieldName}}.Stop()
		`), g)
}
