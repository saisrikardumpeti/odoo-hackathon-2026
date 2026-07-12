package activity_log_repo

import (
 "context"
 "encoding/json"

 "github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
)

func (r *ActivityLogRepository) Create(ctx context.Context, entry models.ActivityLog) error {
 rawMeta, err := json.Marshal(entry.Metadata)
 if err != nil {
  return err
 }
 if rawMeta == nil {
  rawMeta = []byte("{}")
 }

 _, err = r.pool.Exec(ctx,
  `INSERT INTO activity_logs (actor_employee_id, action, entity_type, entity_id, metadata)
   VALUES ($1, $2, $3, $4, $5)`,
  entry.ActorEmployeeID, entry.Action, entry.EntityType, entry.EntityID, rawMeta,
 )
 return err
}
