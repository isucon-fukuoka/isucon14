package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
)

// このAPIをインスタンス内から一定間隔で叩かせることで、椅子とライドをマッチングさせる
func internalGetMatching(w http.ResponseWriter, r *http.Request) {
	fmt.Println("internalGetMatching start!!!")
	ctx := r.Context()
	// MEMO: 一旦最も待たせているリクエストに適当な空いている椅子マッチさせる実装とする。おそらくもっといい方法があるはず…
	ride := &Ride{}
	if err := db.GetContext(ctx, ride, `SELECT * FROM rides WHERE chair_id IS NULL ORDER BY created_at LIMIT 1`); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	fmt.Println("aaaa")
	fmt.Println(ride)
	fmt.Println("bbbb")

	matched := &Chair{}
	if err := db.GetContext(ctx, matched, `SELECT
    cha.*
FROM
    chairs cha
    INNER JOIN chair_locations loc
        ON loc.chair_id = cha.id
WHERE
    cha.is_active = TRUE
    AND (
        SELECT
            count(*) = 0
        FROM
            (
                SELECT
                    count(chair_sent_at) = 6 AS completed
                FROM
                    ride_statuses
                WHERE
                    ride_id IN (SELECT id FROM rides WHERE chair_id = cha.id)
                GROUP BY
                    ride_id
            ) is_completed
        WHERE
            completed = FALSE
    )
ORDER BY
    abs((loc.latitude - ?)) + abs((loc.longitude - ?))
LIMIT
    1`, ride.PickupLatitude, ride.PickupLongitude); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
	}
	fmt.Println("cccc")
	fmt.Println(matched)
	fmt.Println("dddd")

	if _, err := db.ExecContext(ctx, "UPDATE rides SET chair_id = ? WHERE id = ?", matched.ID, ride.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
