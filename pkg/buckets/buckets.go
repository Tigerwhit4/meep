/*
* Package buckets defines the bolt buckets and keys used
 */
package buckets

import (
	"log"

	"github.com/StreamMeBots/meep/pkg/db"
	"github.com/boltdb/bolt"
)

// buckets
var (
	botGreetings          = []byte(`bot.greetings`)
	botStats              = []byte(`bot.stats`)
	userData              = []byte(`user.data`)
	userGreetingTemplates = []byte(`user.greetings.templates`)

	// partial
	userCommands = []byte(`user.commands:`)
)

// Bucket wraps the bolt bucket - future proofing
type Bucket struct {
	*bolt.Bucket
}

func Init() {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(botGreetings); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(botStats); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(userCommands); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(userData); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(userGreetingTemplates); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func BotGreetings(tx *bolt.Tx) Bucket {
	return Bucket{tx.Bucket(botGreetings)}
}

func BotStats(tx *bolt.Tx) Bucket {
	return Bucket{tx.Bucket(botStats)}
}

func UserCommands(userBucket []byte, tx *bolt.Tx) (Bucket, error) {
	bkt, err := tx.CreateBucketIfNotExists(append(userCommands, userBucket...))
	if err != nil {
		return Bucket{}, err
	}
	return Bucket{bkt}, nil
}

func UserData(tx *bolt.Tx) Bucket {
	return Bucket{tx.Bucket(userData)}
}

func UserGreetingTemplates(tx *bolt.Tx) Bucket {
	return Bucket{tx.Bucket(userGreetingTemplates)}
}
