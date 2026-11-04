package consts

import "github.com/gogf/gf/v2/errors/gerror"

var (
	ErrPhoneAlreadyExists              = gerror.New("phone is already registered")
	ErrInvalidCredentials              = gerror.New("invalid phone or password")
	ErrUserDisabled                    = gerror.New("user is disabled")
	ErrAdminInvalidCredentials         = gerror.New("invalid admin username or password")
	ErrAdminDisabled                   = gerror.New("admin user is disabled")
	ErrAdminTokenInvalid               = gerror.New("invalid admin token")
	ErrAdminUserAlreadyExists          = gerror.New("admin user already exists")
	ErrAdminUserNotFound               = gerror.New("admin user not found")
	ErrAdminUserSelfDelete             = gerror.New("cannot delete current admin user")
	ErrCategoryAlreadyExists           = gerror.New("category name already exists under this parent")
	ErrCategoryNotFound                = gerror.New("category not found")
	ErrCategoryParentNotFound          = gerror.New("category parent not found")
	ErrCategoryParentCycle             = gerror.New("category parent would create a cycle")
	ErrCategoryHasChildren             = gerror.New("category has child categories")
	ErrCategoryHasProducts             = gerror.New("category has products")
	ErrAttributeAlreadyExists          = gerror.New("attribute code already exists")
	ErrAttributeNotFound               = gerror.New("attribute not found")
	ErrAttributeHasCategories          = gerror.New("attribute is assigned to categories")
	ErrAttributeHasProducts            = gerror.New("attribute has product values")
	ErrAttributeOptionExists           = gerror.New("attribute option code already exists")
	ErrAttributeOptionNotFound         = gerror.New("attribute option not found")
	ErrAttributeOptionInUse            = gerror.New("attribute option is used by products")
	ErrCategoryAttributeExists         = gerror.New("attribute is already assigned to this category")
	ErrCategoryAttributeNotFound       = gerror.New("category attribute binding not found")
	ErrAttributeGroupExists            = gerror.New("attribute group code or name already exists")
	ErrAttributeGroupNotFound          = gerror.New("attribute group not found")
	ErrAttributeGroupInUse             = gerror.New("attribute group is used by attributes or categories")
	ErrAttributeGroupAttributeNotFound = gerror.New("attribute is not in this group")
	ErrCategoryGroupBindingExists      = gerror.New("attribute group is already bound to this category")
	ErrCategoryGroupBindingNotFound    = gerror.New("attribute group category binding not found")
	ErrTokenInvalid                    = gerror.New("invalid access token")
)

// InvalidParameterError represents a request parameter validation failure.
type InvalidParameterError struct {
	Message string
}

func (e *InvalidParameterError) Error() string {
	return e.Message
}
