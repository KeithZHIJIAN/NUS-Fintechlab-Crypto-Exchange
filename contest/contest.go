package contest

import (
	"fmt"

	"github.com/KeithZHIJIAN/nce-matchingengine/utils"
)

type Participant struct {
	ID      int
	Balance float32
	Asset   float32
}

type Contest struct {
	Symbol         string
	InitialBalance float32
	InitialAsset   int
	InitialPrice   float32
	Participants   []Participant
}

// func main() {
// 	DropTable("PL_WEEKLY_CONTEST_1")
// 	CreateContest("WEEKLY_CONTEST_1")
// 	RegisterToContest(1, 1, "WEEKLY_CONTEST_1", "TEAM_ALICE")
// 	defer utils.DB.Close()
// }

func DropTable(table string) {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
	_, err := utils.DB.Exec(query)
	if err != nil {
		panic(err)
	}
}

func CreateContest(contestName string) {
	_, err := utils.DB.Exec(fmt.Sprintf("CREATE TABLE PL_%s(NAME  VARCHAR(64) UNIQUE);", contestName))
	if err != nil {
		panic(err)
	}
}

func RegisterToContest(contestID, userID int, contestName, participantName string) {
	query := fmt.Sprintf("INSERT INTO PL_%s (NAME) values ($1) ", contestName)
	_, err := utils.DB.Exec(query, participantName)
	if err != nil {
		panic(err)
	}
	query = fmt.Sprintf("INSERT INTO USER_CONTESTS(USERID, CONTESTID, NAME, RETURN, MAXDRAWDOWN) VALUES($1, $2, $3, 10000, 0);")
	_, err = utils.DB.Exec(query, userID, contestID)
	if err != nil {
		panic(err)
	}
}
