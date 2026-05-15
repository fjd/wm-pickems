package scoring

import (
	"encoding/json"
	"log"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// Recompute rebuilds every match_scores and forecast_scores row for all
// in-use configs. Idempotent; cheap at this scale; safe to call on any result
// change (finalize or correction).
func Recompute(app core.App) error {
	configs, _, err := configsInUse(app)
	if err != nil {
		return err
	}

	return app.RunInTransaction(func(tx core.App) error {
		// ---- match scores ----
		msCol, err := tx.FindCollectionByNameOrId("match_scores")
		if err != nil {
			return err
		}
		// Clear and rebuild (small data set).
		old, _ := tx.FindRecordsByFilter("match_scores", "id != ''", "", 0, 0)
		for _, r := range old {
			if err := tx.Delete(r); err != nil {
				return err
			}
		}
		finished, _ := tx.FindRecordsByFilter("matches",
			"finalizedAt != ''", "", 0, 0)
		for _, match := range finished {
			tipList, _ := tx.FindRecordsByFilter("tips",
				"match = {:m}", "", 0, 0, map[string]any{"m": match.Id})
			for _, tip := range tipList {
				for cid, cfg := range configs {
					comp := scoreTip(cfg, match, tip)
					rec := core.NewRecord(msCol)
					rec.Set("user", tip.GetString("user"))
					rec.Set("match", match.Id)
					rec.Set("config", cid)
					rec.Set("points", comp.points())
					cj, _ := json.Marshal(comp)
					rec.Set("components", string(cj))
					if err := tx.Save(rec); err != nil {
						return err
					}
				}
			}
		}

		// ---- forecast scores ----
		fsCol, err := tx.FindCollectionByNameOrId("forecast_scores")
		if err != nil {
			return err
		}
		oldF, _ := tx.FindRecordsByFilter("forecast_scores", "id != ''", "", 0, 0)
		for _, r := range oldF {
			if err := tx.Delete(r); err != nil {
				return err
			}
		}
		forecasts, _ := tx.FindRecordsByFilter("forecasts", "id != ''", "", 0, 0)
		for _, fc := range forecasts {
			for cid, cfg := range configs {
				bd, total := scoreForecast(tx, cfg, fc)
				rec := core.NewRecord(fsCol)
				rec.Set("user", fc.GetString("user"))
				rec.Set("config", cid)
				rec.Set("points", total)
				bj, _ := json.Marshal(bd)
				rec.Set("breakdown", string(bj))
				if err := tx.Save(rec); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// Register wires automatic recompute on result changes and a manual
// superuser trigger.
func Register(app core.App, se *core.ServeEvent) {
	app.OnRecordAfterUpdateSuccess("matches").BindFunc(func(e *core.RecordEvent) error {
		// Recompute when a result is finalized/corrected, or when a knockout
		// match's teams resolve (affects Forecast round scoring).
		if e.Record.GetString("finalizedAt") != "" || e.Record.GetString("stage") != "group" {
			if err := Recompute(e.App); err != nil {
				log.Printf("[scoring] recompute: %v", err)
			}
		}
		return e.Next()
	})

	se.Router.POST("/api/admin/recompute", func(e *core.RequestEvent) error {
		if err := Recompute(app); err != nil {
			return e.JSON(500, map[string]string{"error": err.Error()})
		}
		return e.JSON(200, map[string]string{"status": "ok"})
	}).Bind(apis.RequireSuperuserAuth())
}
