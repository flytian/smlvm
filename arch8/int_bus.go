package arch8

// intBus is an interrupt bus for device
type intBus interface {
	Ncore() byte
	Interrupt(code byte, core byte)
}

// intAllCores generates a interrupt to all cores on the bus.
func intAllCores(bus intBus, code byte) {
	ncore := bus.Ncore()
	for i := byte(0); i < ncore; i++ {
		bus.Interrupt(code, i)
	}
}
