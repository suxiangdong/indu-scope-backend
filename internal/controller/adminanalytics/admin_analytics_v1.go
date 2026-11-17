package adminanalytics

import (
	"context"

	"github.com/suxiangdong/indu-scope-backend/api/adminanalytics/v1"
	analyticsModel "github.com/suxiangdong/indu-scope-backend/internal/model/analytics"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

func (c *ControllerV1) Overview(ctx context.Context, req *v1.OverviewReq) (*v1.OverviewRes, error) {
	days := uint(30)
	if req != nil && req.Days != 0 {
		days = req.Days
	}
	result, err := service.Analytics().Overview(ctx, days)
	if err != nil {
		return nil, err
	}
	return &v1.OverviewRes{OverviewInfo: toAPI(result)}, nil
}

func toAPI(result analyticsModel.Overview) v1.OverviewInfo {
	daily := make([]v1.DailyStat, 0, len(result.Daily))
	for _, item := range result.Daily {
		daily = append(daily, v1.DailyStat{Date: item.Date, ProductClicks: item.ProductClicks, RequirementClicks: item.RequirementClicks, TotalClicks: item.TotalClicks})
	}
	toTop := func(items []analyticsModel.TopContent) []v1.TopContent {
		result := make([]v1.TopContent, 0, len(items))
		for _, item := range items {
			result = append(result, v1.TopContent{ID: item.ID, Name: item.Name, Clicks: item.Clicks})
		}
		return result
	}
	return v1.OverviewInfo{Days: result.Days, StartDate: result.StartDate, TotalClicks: result.TotalClicks, ProductClicks: result.ProductClicks, RequirementClicks: result.RequirementClicks, Daily: daily, TopProducts: toTop(result.TopProducts), TopRequirements: toTop(result.TopRequirements)}
}
