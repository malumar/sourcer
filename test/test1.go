package test


// @Entity(table="user")
type baseUser struct {
}

// @Entity
func Testowa() {

}

// Examples returns the examples found in the files, sorted by Name field.
// The Order fields record the order in which the examples were encountered.
//
// Playable Examples must be in a package whose name ends in "_test".
// An Example is "playable" (the Play field is non-nil) in either of these
// circumstances:
//   - The example function is self-contained: the function references only
//     identifiers from other packages (or predeclared identifiers, such as
//     "int") and the test file does not include a dot import.
//   - The entire test file is the example: the file contains exactly one
//     example function, zero test or benchmark functions, and at least one
//     top-level function, type, variable, or constant declaration other
//     than the example function.


func AA() {

}


// @Title: Pobranie informacji o użytkwoniku
// @Desc: Zwaraca informację o użytkonwniku
// @Desc: Oczywiście o ile, mamy dostęp do tej funkcji
// @Success: 200 {object} string &quot;Success&quot;
// @Failure: 401 {object} Access Denied
// @Failure: 404 {object} Not Found
// @Router: /api/:costam/id [get]
// @Consumes: application/json
// @Produces: application/json
// @ReqGroup: demo,admin
// @ReqRole: administrator
// @AllowGroup: admin, supervisor
// @DisallowGroup: demo
// @AllowRole: nazwa_roli, nazwa_drugiej_roli
// @DisallowRole: moderator
// jezęli zmienne nie wymienione w router, to są traktowane jako parametry opcjonalne w path
func Testowa2(name string, age int) {

}

// @Entity:         ,d211()&21",'z dwukropkiem
func Testowa3() {

}
// EntitaBezMalpy: z dwukropkiem
func Testowa31() {

}
// Entitzx
func Testowa4() {

}
