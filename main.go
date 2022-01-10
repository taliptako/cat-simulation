package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"
)

var session neo4j.Session
var catNames []string

func main() {
	rand.Seed(time.Now().UnixNano())

	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	assignCatNames()
	session = createSession()

	now := time.Now()

	err = createGenesisCats(now)
	if err != nil {
		panic(err)
	}

	for {
		fmt.Println(now.Format("02-January-2006"))

		sixMonthBefore := neo4j.DateOf(now.AddDate(0, -6, 0))

		_, err := session.WriteTransaction(updateBabyCats(sixMonthBefore))
		if err != nil {
			panic(err)
		}

		getAvailableFemaleCatsResult, err := session.ReadTransaction(getAvailableFemaleCats(sixMonthBefore))
		if err != nil {
			panic(err)
		}

		availableCats := getAvailableFemaleCatsResult.([]Cat)

		max := len(availableCats)
		if len(availableCats) > 1 {
			max = max / 2
		}

		for _, f := range availableCats[0:max] {
			maleId, err := session.ReadTransaction(getAvailableMaleCat())
			if err != nil {
				panic(err)
			}

			_, err = session.WriteTransaction(createMatedWithRelation(neo4j.DateOf(now), f.ID, maleId.(int64)))
			if err != nil {
				panic(err)
			}

		}

		twoMonthBefore := now.AddDate(0, -2, 0)
		getReadyToGiveBirthCatsResult, err := session.ReadTransaction(getReadyToGiveBirthCats(neo4j.DateOf(twoMonthBefore)))
		if err != nil {
			panic(err)
		}

		readyToGiveBirthCats := getReadyToGiveBirthCatsResult.([]Cat)

		babyCount := 0
		for _, r := range readyToGiveBirthCats {
			babyCount = giveBirth(now, r) + babyCount
		}

		fmt.Printf("Today %d cat born.\n", babyCount)

		_, err = session.WriteTransaction(makeFemaleCatsAvailable(neo4j.DateOf(now.AddDate(0, -3, 0))))
		if err != nil {
			panic(err)
		}

		now = now.AddDate(0, 0, 1)
	}

}

func createGenesisCats(birthDate time.Time) error {
	adam := Cat{
		Name:      "Adam",
		Gender:    "male",
		BirthDate: neo4j.DateOf(birthDate),
		Status:    "baby",
	}

	eve := Cat{
		Name:      "Eve",
		Gender:    "female",
		BirthDate: neo4j.DateOf(birthDate),
		Status:    "baby",
	}

	_, err := session.WriteTransaction(createCat(adam))
	if err != nil {
		return err
	}
	_, err = session.WriteTransaction(createCat(eve))
	if err != nil {
		return err
	}

	return err
}

func assignCatNames() {
	jsonFile, err := os.Open("cat-names.json")
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully Opened cat-names.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			panic(err)
		}
	}(jsonFile)

	byteValue, _ := ioutil.ReadAll(jsonFile)

	_ = json.Unmarshal([]byte(byteValue), &catNames)
}

func giveBirth(birthDate time.Time, femaleCat Cat) int {
	var babyCatGender string

	babyCount := rand.Intn(7) + 2

	for i := 1; i <= babyCount; i++ {
		if rand.Intn(2) == 0 {
			babyCatGender = "male"
		} else {
			babyCatGender = "female"
		}

		babyCat := Cat{
			Name:      catNames[rand.Intn(len(catNames))],
			Gender:    babyCatGender,
			BirthDate: neo4j.DateOf(birthDate),
			Status:    "baby",
		}

		_, err := session.WriteTransaction(createBabyCat(babyCat, femaleCat))
		if err != nil {
			panic(err)
		}
	}

	return babyCount
}
