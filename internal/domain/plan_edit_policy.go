package domain

// PlanExecutionIdentityChanged reports whether an update rewrites production
// identity after execution has started.
func PlanExecutionIdentityChanged(existing *PlantingPlan, update *PlantingPlanUpsert) bool {
	if existing == nil || update == nil || existing.Status == PlanStatusPlanned {
		return false
	}
	return existing.FieldID != update.FieldID
}
