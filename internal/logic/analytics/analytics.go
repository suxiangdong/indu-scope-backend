package analytics

import (
	"context"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/suxiangdong/indu-scope-backend/internal/consts"
	analyticsModel "github.com/suxiangdong/indu-scope-backend/internal/model/analytics"
	"github.com/suxiangdong/indu-scope-backend/internal/service"
)

type sAnalytics struct{}

var _ service.IAnalytics = (*sAnalytics)(nil)

func New() service.IAnalytics { return &sAnalytics{} }

func init() { service.RegisterAnalytics(New()) }

func (s *sAnalytics) RecordClick(ctx context.Context, userID uint64, contentType analyticsModel.ContentType, contentID uint64) error {
	if userID == 0 {
		return consts.ErrTokenInvalid
	}
	if contentID == 0 {
		return &consts.InvalidParameterError{Message: "content_id is required"}
	}
	if contentType != analyticsModel.ContentTypeProduct && contentType != analyticsModel.ContentTypeRequirement {
		return &consts.InvalidParameterError{Message: "content_type must be product or requirement"}
	}
	table := "products"
	if contentType == analyticsModel.ContentTypeRequirement {
		table = "procurement_requirements"
	}
	count, err := g.DB().Model(table).Where("id", contentID).Count()
	if err != nil {
		return gerror.Wrap(err, "check analytics content")
	}
	if count == 0 {
		return gerror.New("analytics content not found")
	}
	_, err = g.DB().Model("content_click_events").Data(g.Map{
		"content_type": string(contentType),
		"content_id":   contentID,
		"user_id":      userID,
	}).Insert()
	if err != nil {
		return gerror.Wrap(err, "record content click")
	}
	return nil
}

func (s *sAnalytics) Overview(ctx context.Context, days uint) (analyticsModel.Overview, error) {
	if days == 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	start := overviewStart(time.Now(), days)
	base := g.DB().Model("content_click_events").Where("created_at >= ?", start)
	total, err := base.Count()
	if err != nil {
		return analyticsModel.Overview{}, gerror.Wrap(err, "count content clicks")
	}
	productClicks, err := g.DB().Model("content_click_events").Where("content_type", string(analyticsModel.ContentTypeProduct)).Where("created_at >= ?", start).Count()
	if err != nil {
		return analyticsModel.Overview{}, gerror.Wrap(err, "count product clicks")
	}
	requirementClicks, err := g.DB().Model("content_click_events").Where("content_type", string(analyticsModel.ContentTypeRequirement)).Where("created_at >= ?", start).Count()
	if err != nil {
		return analyticsModel.Overview{}, gerror.Wrap(err, "count requirement clicks")
	}

	daily := make([]analyticsModel.DailyStat, 0)
	if err := g.DB().Model("content_click_events").
		Fields("DATE(created_at) AS date, SUM(content_type = 'product') AS product_clicks, SUM(content_type = 'requirement') AS requirement_clicks, COUNT(*) AS total_clicks").
		Where("created_at >= ?", start).Group("DATE(created_at)").OrderAsc("date").Scan(&daily); err != nil {
		return analyticsModel.Overview{}, gerror.Wrap(err, "list daily content clicks")
	}

	topProducts, err := topContent(ctx, "product", "products", start)
	if err != nil {
		return analyticsModel.Overview{}, err
	}
	topRequirements, err := topContent(ctx, "requirement", "procurement_requirements", start)
	if err != nil {
		return analyticsModel.Overview{}, err
	}
	return analyticsModel.Overview{
		Days: days, StartDate: start, TotalClicks: uint64(total), ProductClicks: uint64(productClicks), RequirementClicks: uint64(requirementClicks),
		Daily: daily, TopProducts: topProducts, TopRequirements: topRequirements,
	}, nil
}

// Business reports use China standard time regardless of the host timezone.
// Truncate(24h) rounds to UTC midnight, which is 08:00 in this timezone.
func overviewStart(now time.Time, days uint) time.Time {
	location := time.FixedZone("Asia/Shanghai", 8*60*60)
	local := now.In(location).AddDate(0, 0, -int(days-1))
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, location)
}

func topContent(ctx context.Context, contentType, table string, start time.Time) ([]analyticsModel.TopContent, error) {
	var items []analyticsModel.TopContent
	err := g.DB().Model("content_click_events e").
		LeftJoin(table+" c", "c.id = e.content_id").
		Fields("e.content_id AS id, COALESCE(c.name, '') AS name, COUNT(*) AS clicks").
		Where("e.content_type", contentType).
		Where("e.created_at >= ?", start).
		Group("e.content_id, c.name").OrderDesc("clicks").Limit(10).Scan(&items)
	if err != nil {
		return nil, gerror.Wrap(err, "list top content clicks")
	}
	for index := range items {
		items[index].Name = strings.TrimSpace(items[index].Name)
	}
	return items, nil
}
