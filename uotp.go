package uotp

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrInvalidAccount = errors.New("invalid account")
	ErrInvalidPage    = errors.New("page must be 1 or greater.")
)

type UOTP interface {
	GetSerialNumber() string
	GetAccount() Account

	GenerateToken() string
	SyncTime(ctx context.Context) error
	Issue(ctx context.Context) error
	ResetError(ctx context.Context) error
	GetHistory(ctx context.Context, page int) (*History, error)
	ResetErrorCount(ctx context.Context) error
}

type uotp struct {
	id           string
	oid          uint64
	seed         []byte
	serialNumber string
	timeDiff     int
}

type Account struct {
	ID           string `json:"id"`
	OID          string `json:"oid"`
	Seed         string `json:"seed"`
	SerialNumber string `json:"serial_number"`
	TimeDiff     int    `json:"time_diff"`
}

func New(account *Account) (UOTP, error) {
	var err error

	o := &uotp{}
	if account != nil {
		o.id = account.ID
		o.oid, err = strconv.ParseUint(account.OID, 10, 64)
		if err != nil {
			return nil, ErrInvalidAccount
		}
		o.seed, err = base64.StdEncoding.DecodeString(account.Seed)
		if err != nil {
			return nil, ErrInvalidAccount
		}
		o.serialNumber = fmt.Sprint(account.SerialNumber)
		o.timeDiff = account.TimeDiff
	}

	return o, err
}

func (u *uotp) GetSerialNumber() string {
	return fmt.Sprint(u.serialNumber)
}
func (u *uotp) GetAccount() Account {
	return Account{
		ID:           u.id,
		OID:          strconv.FormatUint(u.oid, 10),
		Seed:         base64.StdEncoding.EncodeToString(u.seed),
		SerialNumber: fmt.Sprint(u.serialNumber),
		TimeDiff:     u.timeDiff,
	}
}

func (otp *uotp) generateToken() string {
	now := uint32(int(otpNow()) + otp.timeDiff)

	time := now / 10
	oid := otp.oid

	accSeed := make([]byte, 11, 11+len(otp.seed))
	timeSeed := make([]byte, 11)

	for i := 10; i >= 0; i-- {
		accSeed[i] = byte(oid) & 0xFF
		timeSeed[i] = byte(time) & 0xFF

		oid >>= 8
		time >>= 8
	}

	accSeed = append(accSeed, otp.seed...)

	h := sha1.New()

	accSeed0 := make([]byte, 64)
	h.Reset()
	h.Write(accSeed)
	copy(accSeed0, h.Sum(nil))

	accSeed1 := make([]byte, 64)
	accSeed2 := make([]byte, 64)
	copy(accSeed1, accSeed0)
	copy(accSeed2, accSeed0)
	for i := 0; i < 64; i++ {
		accSeed1[i] ^= 54
		accSeed2[i] ^= 92
	}

	h.Reset()
	h.Write(accSeed1)
	h.Write(timeSeed)
	digest := h.Sum(nil)

	h.Reset()
	h.Write(accSeed2)
	h.Write(digest)
	digest = h.Sum(nil)

	digit := digest[len(digest)-1] & 0xf

	// noinspection PyShadowingNames
	var token uint32
	token = (token << 8) | uint32(digest[digit+0])
	token = (token << 8) | uint32(digest[digit+1])
	token = (token << 8) | uint32(digest[digit+2])
	token = (token << 8) | uint32(digest[digit+3])
	token &= 0xffffdb

	rem := now % 30
	if 10 <= rem && rem < 20 {
		token |= 4
	} else if 20 <= rem && rem < 30 {
		token |= 32
	}

	return fmt.Sprintf("%07d", token%10000000)
}

func (otp *uotp) GenerateToken() string {
	return humanize(otp.generateToken(), "-", 3, 2)
}

func (u *uotp) SyncTime(ctx context.Context) error {
	now := int(otpNow())

	req := newPacket(opCodeTime)
	resp, err := req.Send(ctx)
	if err != nil {
		return err
	}

	u.timeDiff = resp.payload.(*payloadTime).Time - now
	return nil
}

func (u *uotp) Issue(ctx context.Context) error {
	req := newPacket(opCodeIssue)
	resp, err := req.Send(ctx)
	if err != nil {
		return err
	}

	params := resp.payload.(*payloadIssue)

	u.id = params.userHash
	u.oid = params.oid
	u.seed = params.seed
	u.serialNumber = humanize(params.serialNumber, "-", 4, -1)
	u.timeDiff = 0

	return nil
}

func (u *uotp) ResetError(ctx context.Context) error {
	req := newPacket(opCodeResetErrorCount)
	req.oid = u.oid
	req.setEncryptionInfo(s2b(u.id), u.generateToken())

	_, err := req.Send(ctx)
	return err
}

func (u *uotp) GetHistory(ctx context.Context, page int) (*History, error) {
	if page < 1 {
		return nil, ErrInvalidPage
	}

	req := newPacket(opCodeUseHistory)
	req.oid = u.oid
	req.setEncryptionInfo(s2b(u.id), u.generateToken())

	params := req.payload.(*History)
	params.requestPage = page
	params.requestPeriod = 3

	resp, err := req.Send(ctx)
	if err != nil {
		return nil, err
	}

	return resp.payload.(*History), nil
}

func (u *uotp) ResetErrorCount(ctx context.Context) (err error) {
	req := newPacket(opCodeResetErrorCount)
	req.oid = u.oid
	req.setEncryptionInfo(s2b(u.id), u.generateToken())

	_, err = req.Send(ctx)
	return
}
