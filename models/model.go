// common
package models

import (
	"labix.org/v2/mgo"
	"log"
)

const (
	dbName   = "cmdcoin"
	cBlocks  = "blocks"
	cTxs     = "txs"
	cOutputs = "outputs"
	cWallets = "wallets"
)

var (
	mgoSession *mgo.Session
)

func getSession() *mgo.Session {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.Dial("localhost")
		if err != nil {
			log.Println(err) // no, not really
		}
	}
	return mgoSession.Clone()
}

func withCollection(collection string, safe *mgo.Safe, s func(*mgo.Collection) error) error {
	session := getSession()
	defer session.Close()

	session.SetSafe(safe)
	c := session.DB(dbName).C(collection)
	return s(c)
}

func exists(collection string, query interface{}) (bool, error) {
	b := false
	q := func(c *mgo.Collection) error {
		n, err := c.Find(query).Count()
		b = n > 0
		return err
	}

	err := withCollection(collection, nil, q)
	return b, err
}

func find(collection string, query interface{}, selector interface{},
	skip, limit int, sortFields []string, total *int, result interface{}) error {
	q := func(c *mgo.Collection) error {
		qy := c.Find(query)
		var err error

		if selector != nil {
			qy = qy.Select(selector)
		}

		if total != nil {
			if *total, err = qy.Count(); err != nil {
				return err
			}
		}

		if result == nil {
			return err
		}

		if limit > 0 {
			qy = qy.Limit(limit)
		}
		if skip > 0 {
			qy = qy.Skip(skip)
		}
		if len(sortFields) > 0 {
			qy = qy.Sort(sortFields...)
		}

		return qy.All(result)
	}

	return withCollection(collection, nil, q)
}

func findOne(collection string, query interface{}, sortFields []string, result interface{}) error {
	q := func(c *mgo.Collection) error {
		var err error
		qy := c.Find(query)

		if result == nil {
			return err
		}

		if len(sortFields) > 0 {
			qy = qy.Sort(sortFields...)
		}

		return qy.One(result)
	}

	return withCollection(collection, nil, q)
}

func save(collection string, o interface{}, safe bool) error {
	insert := func(c *mgo.Collection) error {
		return c.Insert(o)
	}

	if safe {
		return withCollection(collection, &mgo.Safe{}, insert)
	}
	return withCollection(collection, nil, insert)
}

func update(collection string, selector, change interface{}, safe bool) error {
	update := func(c *mgo.Collection) error {
		return c.Update(selector, change)
	}
	if safe {
		return withCollection(collection, &mgo.Safe{}, update)
	}
	return withCollection(collection, nil, update)
}

func upsert(collection string, selector, change interface{}, safe bool) (info *mgo.ChangeInfo, err error) {
	upsert := func(c *mgo.Collection) (err error) {
		info, err = c.Upsert(selector, change)
		//log.Println(chinfo, err)
		return err
	}
	if safe {
		err = withCollection(collection, &mgo.Safe{}, upsert)
		return
	}
	err = withCollection(collection, nil, upsert)
	return
}

func remove(collection string, selector interface{}, safe bool) error {
	rm := func(c *mgo.Collection) error {
		err := c.Remove(selector)
		if err == mgo.ErrNotFound {
			return nil
		}
		return err
	}
	if safe {
		return withCollection(collection, &mgo.Safe{}, rm)
	}
	return withCollection(collection, nil, rm)
}

func apply(collection string, selector interface{}, change mgo.Change, result interface{}) (info *mgo.ChangeInfo, err error) {
	apply := func(c *mgo.Collection) (err error) {
		info, err = c.Find(selector).Apply(change, result)
		return err
	}

	err = withCollection(collection, nil, apply)
	return
}

func ensureIndex(collection string, keys ...string) error {
	ensure := func(c *mgo.Collection) error {
		return c.EnsureIndexKey(keys...)
	}

	return withCollection(collection, nil, ensure)
}
