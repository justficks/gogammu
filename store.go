package gammu

import (
	"sync"
)

type Store struct {
	mu     sync.RWMutex
	modems map[int]*Modem
}

var instance *Store
var once sync.Once

func GetStore() *Store {
	once.Do(func() {
		instance = &Store{
			modems: make(map[int]*Modem),
		}
	})
	return instance
}

func (s *Store) AddModem(m *Modem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modems[m.Num] = m
}

func (s *Store) DeleteModem(num int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.modems, num)
}

func (s *Store) GetModem(num int) (*Modem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	modem, ok := s.modems[num]
	return modem, ok
}

func (s *Store) GetModemByIMSI(imsi string) (*Modem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, modem := range s.modems {
		if modem.IMSI == imsi {
			return modem, true
		}
	}
	return nil, false
}

func (s *Store) GetModems() map[int]*Modem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modems
}

func (s *Store) GetModemsByStatus(status ModemStatus) map[int]*Modem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[int]*Modem)
	for id, modem := range s.modems {
		if modem.Status == status {
			result[id] = modem
		}
	}
	return result
}

func (s *Store) SetModemsIdentify(modems []ModemIdentify) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, i := range modems {
		if modem, ok := s.modems[i.ModemNumber]; ok {

			if i.Error != "" {
				if i.ErrorCode == 114 {
					modem.Status = NoSIM
					modem.Error = i.Error
					continue
				}
				modem.Status = Error
				modem.Error = i.Error
				continue
			}

			modem.IMEI = i.IMEI
			modem.IMSI = i.IMSI
			modem.Device = i.Device
			modem.Manufacturer = i.Manufacturer
			modem.Model = i.Model
			modem.Firmware = i.Firmware
			modem.Status = Stop
		}
	}
}

func (s *Store) SetModemsNetwork(modems []ModemNetwork) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, i := range modems {
		if modem, ok := s.modems[i.ModemNumber]; ok {
			if i.Error != "" {
				modem.Status = Error
				modem.Error = i.Error
				continue
			}

			modem.NetworkState = i.NetworkState
			modem.Network = i.Network
			modem.NameInPhone = i.NameInPhone
			modem.PacketNetworkState = i.PacketNetworkState
			modem.PacketNetwork = i.PacketNetwork
			modem.GPRS = i.GPRS
			modem.Status = Stop
		}
	}
}

func (s *Store) SetModemIdentify(mIdent *ModemIdentify) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if modem, ok := s.modems[mIdent.ModemNumber]; ok {
		modem.IMEI = mIdent.IMEI
		modem.IMSI = mIdent.IMSI
		modem.Device = mIdent.Device
		modem.Manufacturer = mIdent.Manufacturer
		modem.Model = mIdent.Model
		modem.Firmware = mIdent.Firmware
	} else {
		s.modems[mIdent.ModemNumber] = &Modem{
			Num:          mIdent.ModemNumber,
			IMEI:         mIdent.IMEI,
			IMSI:         mIdent.IMSI,
			Device:       mIdent.Device,
			Manufacturer: mIdent.Manufacturer,
			Model:        mIdent.Model,
			Firmware:     mIdent.Firmware,
			Status:       Stop,
		}
	}

}

func (s *Store) SetModemsDetect(devices map[int]string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for modemNumber, device := range devices {
		if modem, ok := s.modems[modemNumber]; ok {
			modem.Device = device
		} else {
			s.modems[modemNumber] = &Modem{
				Num:    modemNumber,
				Device: device,
			}
		}
	}
}

func (s *Store) SetModemMonitor(n int, m *ModemMonitor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if modem, ok := s.modems[n]; ok {
		modem.PhoneID = m.PhoneID
		modem.IMEI = m.IMEI
		modem.IMSI = m.IMSI
		modem.Sent = m.Sent
		modem.Received = m.Received
		modem.Failed = m.Failed
		modem.BatterPercent = m.BatterPercent
		modem.NetworkSignal = m.NetworkSignal
	}
}

func (s *Store) SetModemsMonitor(modems []ModemMonitor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, i := range modems {
		if modem, ok := s.modems[i.ModemNumber]; ok {
			if i.Error != "" {
				modem.Status = Error
				modem.Error = i.Error
				continue
			}

			modem.PhoneID = i.PhoneID
			modem.IMEI = i.IMEI
			modem.IMSI = i.IMSI
			modem.Sent = i.Sent
			modem.Received = i.Received
			modem.Failed = i.Failed
			modem.BatterPercent = i.BatterPercent
			modem.NetworkSignal = i.NetworkSignal
		}
	}
}

func (s *Store) SetModemsRun(modems []ModemRun) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, i := range modems {
		if modem, ok := s.modems[i.ModemNumber]; ok {
			if i.Run {
				modem.Status = Run
			} else {
				modem.Status = Error
				modem.Error = i.Error
			}
		}
	}
}

func (s *Store) SetModemNetwork(n int, m *ModemNetwork) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if modem, ok := s.modems[n]; ok {
		modem.NetworkState = m.NetworkState
		modem.Network = m.Network
		modem.NameInPhone = m.NameInPhone
		modem.PacketNetworkState = m.PacketNetworkState
		modem.PacketNetwork = m.PacketNetwork
		modem.GPRS = m.GPRS
	}
}

func (s *Store) SetModemsStop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, i := range s.modems {
		if i.Status == Run {
			i.Status = Stop
		}
	}
}

func (s *Store) SetModemStatus(num int, status ModemStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if modem, ok := s.modems[num]; ok {
		modem.Status = status
	}
}

func (s *Store) SetModemError(num int, error string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if modem, ok := s.modems[num]; ok {
		modem.Status = Error
		modem.Error = error
	}
}

func (s *Store) SetModemPID(num int, pid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if modem, ok := s.modems[num]; ok {
		modem.PID = pid
	}
}

func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modems = make(map[int]*Modem)
}
