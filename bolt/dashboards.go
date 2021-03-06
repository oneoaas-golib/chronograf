package bolt

import (
	"context"
	"strconv"

	"github.com/boltdb/bolt"
	"github.com/influxdata/chronograf"
	"github.com/influxdata/chronograf/bolt/internal"
)

// Ensure DashboardsStore implements chronograf.DashboardsStore.
var _ chronograf.DashboardsStore = &DashboardsStore{}

// DashboardBucket is the bolt bucket dashboards are stored in
var DashboardBucket = []byte("Dashoard")

// DashboardsStore is the bolt implementation of storing dashboards
type DashboardsStore struct {
	client *Client
	IDs    chronograf.DashboardID
}

// All returns all known dashboards
func (d *DashboardsStore) All(ctx context.Context) ([]chronograf.Dashboard, error) {
	var srcs []chronograf.Dashboard
	if err := d.client.db.View(func(tx *bolt.Tx) error {
		if err := tx.Bucket(DashboardBucket).ForEach(func(k, v []byte) error {
			var src chronograf.Dashboard
			if err := internal.UnmarshalDashboard(v, &src); err != nil {
				return err
			}
			srcs = append(srcs, src)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return srcs, nil
}

// Add creates a new Dashboard in the DashboardsStore
func (d *DashboardsStore) Add(ctx context.Context, src chronograf.Dashboard) (chronograf.Dashboard, error) {
	if err := d.client.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(DashboardBucket)
		id, _ := b.NextSequence()

		src.ID = chronograf.DashboardID(id)
		strID := strconv.Itoa(int(id))
		if v, err := internal.MarshalDashboard(src); err != nil {
			return err
		} else if err := b.Put([]byte(strID), v); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return chronograf.Dashboard{}, err
	}

	return src, nil
}

// Get returns a Dashboard if the id exists.
func (d *DashboardsStore) Get(ctx context.Context, id chronograf.DashboardID) (chronograf.Dashboard, error) {
	var src chronograf.Dashboard
	if err := d.client.db.View(func(tx *bolt.Tx) error {
		strID := strconv.Itoa(int(id))
		if v := tx.Bucket(DashboardBucket).Get([]byte(strID)); v == nil {
			return chronograf.ErrDashboardNotFound
		} else if err := internal.UnmarshalDashboard(v, &src); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return chronograf.Dashboard{}, err
	}

	return src, nil
}

// Delete the dashboard from DashboardsStore
func (d *DashboardsStore) Delete(ctx context.Context, dash chronograf.Dashboard) error {
	if err := d.client.db.Update(func(tx *bolt.Tx) error {
		if err := tx.Bucket(DashboardBucket).Delete(itob(int(dash.ID))); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

// Update the dashboard in DashboardsStore
func (d *DashboardsStore) Update(ctx context.Context, dash chronograf.Dashboard) error {
	if err := d.client.db.Update(func(tx *bolt.Tx) error {
		// Get an existing dashboard with the same ID.
		b := tx.Bucket(DashboardBucket)
		strID := strconv.Itoa(int(dash.ID))
		if v := b.Get([]byte(strID)); v == nil {
			return chronograf.ErrDashboardNotFound
		}

		if v, err := internal.MarshalDashboard(dash); err != nil {
			return err
		} else if err := b.Put([]byte(strID), v); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
