package simulation

import (
    "fmt"
    "UTC_IA04/pkg/models"
)

type Centrale struct {
    Map map[models.Position]Status
}

type Status struct {
    PeopleCount int
    DistressCount int
}

func (c *Centrale) SendInfoToDrone(id int) {
    fmt.Printf("Sending info to drone %d\n", id)
}
