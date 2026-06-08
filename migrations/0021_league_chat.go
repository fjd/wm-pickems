package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

// League chat (v1): a per-league text chat for private leagues.
//
//   - league_messages: one row per message. Reads are gated to league members
//     of a non-Global league via a back-relation rule, which is what authorises
//     the realtime (SSE) subscription clients use for live updates. Writes have
//     no collection rule — they only happen through the validated chat endpoints
//     (internal/chat), whose app.Save still fires the realtime event.
//   - league_reads: a per-(user,league) last-read marker driving unread badges.
//     Managed entirely server-side, so no client rules.
const (
	nLeagueMessages = "league_messages"
	nLeagueReads    = "league_reads"
)

func init() {
	m.Register(func(app core.App) error {
		leagues, err := app.FindCollectionByNameOrId("leagues")
		if err != nil {
			return err
		}
		users, err := app.FindCollectionByNameOrId("users")
		if err != nil {
			return err
		}

		// ---- league_messages ----
		if _, err := app.FindCollectionByNameOrId(nLeagueMessages); err != nil {
			msg := core.NewBaseCollection(nLeagueMessages)
			// Members of a non-Global league may read (and thus realtime-subscribe).
			read := "@request.auth.id != '' && " +
				"league.inviteCode != 'GLOBAL' && " +
				"league.league_members_via_league.user ?= @request.auth.id"
			msg.ListRule = &read
			msg.ViewRule = &read
			// Create/update/delete intentionally nil = only the chat endpoints write.
			msg.Fields.Add(&core.RelationField{Name: "league", CollectionId: leagues.Id, MaxSelect: 1, Required: true, CascadeDelete: true})
			msg.Fields.Add(&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1, Required: true, CascadeDelete: true})
			msg.Fields.Add(&core.TextField{Name: "text", Required: true, Max: 2000})
			msg.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
			msg.AddIndex("idx_msg_league_created", false, "league, created", "")
			if err := app.Save(msg); err != nil {
				return err
			}
		}

		// ---- league_reads ----
		if _, err := app.FindCollectionByNameOrId(nLeagueReads); err != nil {
			rd := core.NewBaseCollection(nLeagueReads)
			// Server-managed only (no client rules).
			rd.Fields.Add(&core.RelationField{Name: "league", CollectionId: leagues.Id, MaxSelect: 1, Required: true, CascadeDelete: true})
			rd.Fields.Add(&core.RelationField{Name: "user", CollectionId: users.Id, MaxSelect: 1, Required: true, CascadeDelete: true})
			rd.Fields.Add(&core.DateField{Name: "lastRead"})
			rd.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})
			rd.AddIndex("idx_reads_league_user", true, "league, user", "")
			if err := app.Save(rd); err != nil {
				return err
			}
		}
		return nil
	}, func(app core.App) error {
		for _, name := range []string{nLeagueMessages, nLeagueReads} {
			if c, err := app.FindCollectionByNameOrId(name); err == nil {
				if err := app.Delete(c); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
