package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/boltdb/bolt"
	"github.com/pspiagicw/colorlog"
)

type Package struct {
	Name      string    `json:"name"`
	Checksum  string    `json:"checkSum"`
	Version   string    `json:"version"`
	Path      string    `json:"path"`
	Installed time.Time `json:"installed"`
}

type Plugin struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Path        string    `json:"path"`
	Installed   time.Time `json:"installed"`
}
type GroomDatabase struct {
	Packages map[string]*Package `json:"packages"`
	Plugins  map[string]*Plugin  `json:"plugins"`
}

func getDatabasePath() (string, error) {
	location, exists := os.LookupEnv("XDG_DATA_HOME")
	if !exists {
		colorlog.LogInfo("Not using $XDG_DATA_HOME, env variable not present")
		homedir, err := os.UserHomeDir()
		if err != nil {
			colorlog.LogError("Error while getting $HOME directory: %v", err)
			return "", fmt.Errorf("Error while getting $HOME directory: %v", err)
		}
		d := filepath.Join(homedir, ".local")
		d = filepath.Join(d, "share")
		d = filepath.Join(d, "groom")
		d = filepath.Join(d, "db")
		colorlog.LogInfo("Using %s for database", d)
		return d, nil
	}
	d := filepath.Join(location, "groom")
	d = filepath.Join(d, "db")
	colorlog.LogInfo("Using %s for database", d)
	return d, nil

}
func ParseDatabase() (*GroomDatabase, error) {
	path, err := getDatabasePath()

	if err != nil {
		return nil, err
	}

	database, err := readDatabase(path)

	if err != nil {
		return nil, err
	}
	return database, nil

}
func readDatabase(path string) (*GroomDatabase, error) {
	gdb := new(GroomDatabase)
	gdb.Packages = map[string]*Package{}
	gdb.Plugins = map[string]*Plugin{}
	db, err := bolt.Open(path, 666, nil)
	defer db.Close()
	if err != nil {
		return nil, fmt.Errorf("Error reading the database: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("packages"))
		if err != nil {
			return fmt.Errorf("Error opening bucket: %v", err)
		}
		if b == nil {
			return fmt.Errorf("Error opening bucket: %v", err)
		}
		c := b.Cursor()

		for key, value := c.First(); key != nil; key, value = c.Next() {
			p := new(Package)
			err := json.Unmarshal(value, &p)
			if err != nil {
				return fmt.Errorf("Error Unmarshal struct: %v", err)
			}
			gdb.Packages[string(key)] = p
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Error reading database: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("plugins"))
		if err != nil {
			return fmt.Errorf("Error opening bucket: %v", err)
		}
		c := b.Cursor()

		for key, value := c.First(); key != nil; key, value = c.Next() {
			p := new(Plugin)
			err := json.Unmarshal(value, &p)
			if err != nil {
				return fmt.Errorf("Error Unmarshal struct: %v", err)
			}
			gdb.Plugins[string(key)] = p
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("Error reading database: %v", err)
	}

	return gdb, nil
}
func AddPackage(pack Package) error {
	path, err := getDatabasePath()

	if err != nil {
		return err
	}

	db, err := bolt.Open(path, 0600, nil)
	defer db.Close()
	if err != nil {
		return fmt.Errorf("Error reading the database: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("packages"))
		if err != nil || b == nil {
			return fmt.Errorf("Error creating bucket: %v", err)
		}
		contents, err := json.Marshal(pack)
		if err != nil {
			return fmt.Errorf("Error marshalling pacakge: %v", err)
		}
		err = b.Put([]byte(pack.Name), contents)
		if err != nil {
			return fmt.Errorf("Error inserting into database: %v", err)
		}
		return nil

	})
	return err
}
func AddPlugin(plugin Plugin) error {
	path, err := getDatabasePath()

	if err != nil {
		return err
	}

	db, err := bolt.Open(path, 0600, nil)
	defer db.Close()
	if err != nil {
		return fmt.Errorf("Error reading the database: %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("plugins"))
		if err != nil || b == nil {
			return fmt.Errorf("Error creating bucket: %v", err)
		}
		contents, err := json.Marshal(plugin)
		if err != nil {
			return fmt.Errorf("Error marshalling pacakge: %v", err)
		}
		err = b.Put([]byte(plugin.Name), contents)
		if err != nil {
			return fmt.Errorf("Error inserting into database: %v", err)
		}
		return nil

	})
	return err
}
