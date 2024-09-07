package z 


import (

	"github.com/charmbracelet/huh"

)

var startNow bool 
var oldProject bool


func createInitialForm() *huh.Form {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Title("Start tracking the Task now?").Value(&startNow),
			huh.NewConfirm().Title("Track this on an already existing Project?").Value(&oldProject),
			),
	)
}

func createMainForm() *huh.Form {
	if oldProject{
		projectForm := 
	form := huh.NewForm(
		huh.NewGroup(
			
		)
}
		huh.NewGroup(
			
			)



		)
}

