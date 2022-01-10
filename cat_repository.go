package main

import (
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j/dbtype"
)

func createCat(cat Cat) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run("CREATE (c:Cat {name: $name, gender: $gender, birth_date: $birth_date, status: $status})",
			map[string]interface{}{"name": cat.Name, "gender": cat.Gender, "birth_date": cat.BirthDate, "status": cat.Status})

		if err != nil {
			return nil, err
		}

		return result.Consume()
	}
}

func getAvailableFemaleCats(birthDate dbtype.Date) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run("MATCH (c:Cat {gender: $gender, status: $status}) "+
			"WHERE c.birth_date < $birth_date RETURN id(c), c.name, c.gender, c.birth_date",
			map[string]interface{}{"birth_date": birthDate, "gender": "female", "status": "available"})

		if err != nil {
			return nil, err
		}

		var cats []Cat

		for result.Next() {

			id, _ := result.Record().Get("id(c)")
			name, _ := result.Record().Get("c.name")
			gender, _ := result.Record().Get("c.gender")
			birthDate, _ := result.Record().Get("c.birth_date")

			cats = append(cats, Cat{
				ID:        id.(int64),
				Name:      name.(string),
				Gender:    gender.(string),
				BirthDate: birthDate.(dbtype.Date),
			})
		}

		return cats, nil
	}
}

func getAvailableMaleCat() neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run("MATCH (c:Cat {gender: $gender, status: $status}) WITH c ORDER BY rand() RETURN id(c) LIMIT 1",
			map[string]interface{}{"gender": "male", "status": "available"})

		if err != nil {
			return nil, err
		}

		single, _ := result.Single()

		maleId, _ := single.Get("id(c)")

		maleId = maleId.(int64)

		return maleId, nil
	}
}

func updateBabyCats(birthDate dbtype.Date) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run("MATCH (c:Cat {status: $status}) WHERE (c.birth_date < $birth_date) SET c.status = $new_status RETURN c.name, c.gender, c.birth_date",
			map[string]interface{}{"birth_date": birthDate, "status": "baby", "new_status": "available"})

		if err != nil {
			return nil, err
		}

		return result.Consume()
	}
}

func createMatedWithRelation(mateDate dbtype.Date, femaleId int64, maleId int64) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run("MATCH (f:Cat) WHERE (id(f) = $female_id)"+
			"MATCH (m:Cat) WHERE (id(m) = $maleId)"+
			"MERGE (f)-[:MATED_WITH]->(m)"+
			"SET f.last_mated_id = id(m), f.last_mated_at = $last_mated_at, m.last_mated_at = $last_mated_at, f.status = $pregnant",
			map[string]interface{}{"female_id": femaleId, "maleId": maleId, "available": "available", "pregnant": "pregnant", "last_mated_at": mateDate})

		if err != nil {
			return nil, err
		}

		return result.Consume()
	}
}

func getReadyToGiveBirthCats(lastMatedAt dbtype.Date) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run("MATCH (c:Cat {gender: $gender}) "+
			"WHERE (c.last_mated_at = $last_mated_at) RETURN id(c), c.name, c.gender, c.birth_date, c.last_mated_id",
			map[string]interface{}{"last_mated_at": lastMatedAt, "gender": "female"})

		if err != nil {
			return nil, err
		}

		var cats []Cat

		for result.Next() {

			id, _ := result.Record().Get("id(c)")
			name, _ := result.Record().Get("c.name")
			gender, _ := result.Record().Get("c.gender")
			birthDate, _ := result.Record().Get("c.birth_date")
			lastMatedId, _ := result.Record().Get("c.last_mated_id")

			cats = append(cats, Cat{
				ID:          id.(int64),
				Name:        name.(string),
				Gender:      gender.(string),
				BirthDate:   birthDate.(dbtype.Date),
				LastMatedId: lastMatedId.(int64),
			})
		}

		return cats, nil
	}
}

func createBabyCat(babyCat Cat, femaleCat Cat) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run("MATCH (f:Cat) WHERE (id(f) = $female_id) MATCH (m:Cat) WHERE (id(m) = $male_id)"+
			"CREATE (b:Cat {name: $name, gender: $gender, birth_date: $birth_date, status: $status})"+
			"MERGE (b)-[r:CHILD_OF]->(f) MERGE (b)-[t:CHILD_OF]->(m)",
			map[string]interface{}{"female_id": femaleCat.ID, "male_id": femaleCat.LastMatedId, "name": babyCat.Name, "gender": babyCat.Gender, "birth_date": babyCat.BirthDate, "status": babyCat.Status})

		if err != nil {
			return nil, err
		}

		return result.Consume()
	}
}

func makeFemaleCatsAvailable(lastMatedAt dbtype.Date) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		var result, err = tx.Run("MATCH (f:Cat) WHERE (f.last_mated_at = $last_mated_at) SET f.status = $status",
			map[string]interface{}{"last_mated_at": lastMatedAt, "status": "available"})

		if err != nil {
			return nil, err
		}

		return result.Consume()
	}
}
