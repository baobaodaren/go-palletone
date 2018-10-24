package jury

import "fmt"

func GetJurors() ([]Processor, error) {
	//todo tmp
	var jurors []Processor
	var juror Processor
	juror.ptype = TJury

	for i := 0; i < 10; i++ {
		fmt.Sprintf(juror.name, "juror_%d", i)
		jurors = append(jurors, juror)
	}

	return jurors, nil
}
