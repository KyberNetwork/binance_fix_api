package fix

import (
	"context"

	"github.com/google/uuid"
	"github.com/quickfixgo/field"
	"github.com/quickfixgo/quickfix"
)

/*
Tag         Name                            Type            Required
6136        ReqID                           STRING          Y
25003       NoLimitIndicators               NUM_IN_GROUP    Y
» 25004     LimitType                       CHAR            Y (1: ORDER_LIMIT, 2: MESSAGE_LIMIT)
» 25005     LimitCount                      INT             Y
» 25006     LimitMax                        INT             Y
» 25007     LimitResetInterval              INT             N
» 25008     LimitResetIntervalResolution    CHAR            N (s: SECOND, m: MINUTE, h: HOUR, d: DAY)
*/
type LimitType string

const (
	LimitTypeOrder   LimitType = "1"
	LimitTypeMessage LimitType = "2"
)

type LimitResolution string

const (
	LimitResolutionSecond LimitResolution = "s"
	LimitResolutionMinute LimitResolution = "m"
	LimitResolutionHour   LimitResolution = "h"
	LimitResolutionDay    LimitResolution = "d"
)

type Limit struct {
	LimitType                    LimitType
	LimitCount                   int
	LimitMax                     int
	LimitResetInterval           int
	LimitResetIntervalResolution LimitResolution
}

type LimitResponse struct {
	ReqID             string
	NoLimitIndicators int
	Limits            []Limit
}

type LimitService struct {
	c *Client
}

func (c *Client) NewGetLimitService() *LimitService {
	return &LimitService{c}
}

func (s *LimitService) Do(ctx context.Context) (LimitResponse, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return LimitResponse{}, err
	}

	msg := quickfix.NewMessage()
	msg.Header.Set(field.NewMsgType(msgType_LIMIT_REQUEST))

	msg.Body.SetString(tagGetLimitReqID, id.String())

	resp, err := s.c.Call(ctx, id.String(), msg)
	if err != nil {
		return LimitResponse{}, err
	}

	// parse response
	reqID, err := resp.Body.GetString(tagGetLimitReqID)
	if err != nil {
		return LimitResponse{}, err
	}

	noLimitIndicators, err := resp.Body.GetInt(tagNoLimitIndicators)
	if err != nil {
		return LimitResponse{}, err
	}

	f := quickfix.NewRepeatingGroup(
		tagNoLimitIndicators, quickfix.GroupTemplate{
			quickfix.GroupElement(tagLimitType),
			quickfix.GroupElement(tagLimitCount),
			quickfix.GroupElement(tagLimitMax),
			quickfix.GroupElement(tagLimitResetInterval),
			quickfix.GroupElement(tagLimitResetIntervalResolution),
		})
	err = resp.Body.GetGroup(f)
	if err != nil {
		return LimitResponse{}, err
	}

	limits := make([]Limit, 0)
	for i := range f.Len() {
		limit := f.Get(i)

		limitType, err := limit.GetString(tagLimitType)
		if err != nil {
			return LimitResponse{}, err
		}

		limitCount, err := limit.GetInt(tagLimitCount)
		if err != nil {
			return LimitResponse{}, err
		}

		limitMax, err := limit.GetInt(tagLimitMax)
		if err != nil {
			return LimitResponse{}, err
		}

		var (
			limitResetInterval           int
			limitResetIntervalResolution string
		)
		if limit.Has(tagLimitResetInterval) {
			limitResetInterval, err = limit.GetInt(tagLimitResetInterval)
			if err != nil {
				return LimitResponse{}, err
			}
		}
		if limit.Has(tagLimitResetIntervalResolution) {
			limitResetIntervalResolution, err = limit.GetString(tagLimitResetIntervalResolution)
			if err != nil {
				return LimitResponse{}, err
			}
		}

		limits = append(limits, Limit{
			LimitType:                    LimitType(limitType),
			LimitCount:                   limitCount,
			LimitMax:                     limitMax,
			LimitResetInterval:           limitResetInterval,
			LimitResetIntervalResolution: LimitResolution(limitResetIntervalResolution),
		})
	}

	return LimitResponse{
		ReqID:             reqID,
		NoLimitIndicators: noLimitIndicators,
		Limits:            limits,
	}, nil
}
