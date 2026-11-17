-- 小程序首次发布：用户协议、隐私政策和普通会员协议。
-- 由运营方填写并手动执行；此文件不会随服务启动或数据库迁移自动执行。
-- 先部署支持 user/privacy 协议的后端，确保已完成 000038 及后续迁移。
-- 在业务库中、同一连接内执行整个文件。任一语句报错应 ROLLBACK，停止执行，勿使用 mysql --force。
-- 主体及联系邮箱由运营方提供；保存规则采用最短必要原则，不设置无依据的统一天数。
-- 执行前核实存储地域、服务提供方及保留规则与实际运营一致；可选电话/地址/微信留空时停止展示。
-- 如仍有 __...__ 占位值或必要主体/邮箱为空，不会写入任何协议。
-- 本次按新版本追加四类协议，旧行、旧版本以及订单绑定的协议号均保持原样。
-- 参考：微信隐私指南 https://developers.weixin.qq.com/miniprogram/dev/framework/user-privacy/PrivacyAuthorize.html
-- 参考：个人信息保护法 https://www.miit.gov.cn/jgsj/zfs/fl/art/2022/art_515a4b20c12f430eab54bb4f56d89f56.html

SET NAMES utf8mb4;

SET @operator_name = '北京乘风扶摇科技有限公司';
SET @operator_address = '';
SET @contact_phone = '';
SET @contact_email = 'suxiangd@gmail.com';
SET @service_hours = '';
SET @wechat_account = '';
SET @account_retention = '在账号存续及提供账号、内容展示和供需对接服务所必要的期间保留；处理目的已实现、不再必要，或者注销、撤回同意等依法应予删除的情形发生后，依法删除或匿名化';
SET @verification_retention = '在完成认证核验、维持认证有效状态及处理必要申诉所需的期间保留；核验及相关处理目的已达成，或者注销、撤回相应同意等依法应予删除的情形发生后，依法删除或匿名化';
SET @order_retention = '按适用法律规定及履行会员服务、支付核对、退款和必要争议处理所需期间保留；相关法定及必要处理期限届满后，依法删除或匿名化';
SET @analytics_log_retention = '仅在实现产品需求访问统计、运行维护、故障排查及安全保护所必要的最短期间保留；处理目的达成或不再必要后删除或匿名化，适用法律另有规定的按其规定执行';
SET @storage_region = '中国境内';
SET @request_response_time = '核验身份后及时处理，并告知处理结果；依法不能满足请求的，将说明理由';
-- 以下列明当前实际产品服务及处理范围，不虚构未知的服务商法定主体名称。
SET @service_providers = '微信平台：在用户主动授权时提供微信登录标识、手机号及微信隐私授权能力，用于注册登录和身份识别。微信支付：处理会员订单对应的支付标识、金额及必要支付退款信息，向平台反馈支付或退款状态，用于完成收付款及权益核对。阿里云 OSS：保存用户主动上传的产品和需求图片、认证资料等对象，供平台按权限及临时访问链接提供业务所需访问。阿里云 OCR：识别用户主动提交的营业执照图片并返回识别字段，用于减少认证填写和辅助核验。';
SET @effective_date = DATE_FORMAT(CURRENT_DATE(), '%Y-%m-%d');
SET @nl = CHAR(10, 10);
SET @contact_details = CONCAT('联系邮箱：', @contact_email,
    IF(CHAR_LENGTH(TRIM(@contact_phone)) > 0, CONCAT('；联系电话：', @contact_phone), ''),
    IF(CHAR_LENGTH(TRIM(@operator_address)) > 0, CONCAT('；联系地址：', @operator_address), ''), '。');

SET @release_inputs_ready = (
    CHAR_LENGTH(TRIM(@operator_name)) > 0
    AND CHAR_LENGTH(TRIM(@contact_email)) > 0
    AND CHAR_LENGTH(TRIM(@account_retention)) > 0
    AND CHAR_LENGTH(TRIM(@verification_retention)) > 0
    AND CHAR_LENGTH(TRIM(@order_retention)) > 0
    AND CHAR_LENGTH(TRIM(@analytics_log_retention)) > 0
    AND CHAR_LENGTH(TRIM(@storage_region)) > 0
    AND CHAR_LENGTH(TRIM(@request_response_time)) > 0
    AND CHAR_LENGTH(TRIM(@service_providers)) > 0
    AND LOCATE('__', CONCAT_WS('', @operator_name, @operator_address, @contact_phone, @contact_email,
        @service_hours, @wechat_account, @account_retention, @verification_retention,
        @order_retention, @analytics_log_retention, @storage_region, @request_response_time, @service_providers)) = 0
);

SET @user_content = CONCAT(
    '源仓用户协议', @nl, '生效日期：', @effective_date, @nl,
    '一、服务提供者', @nl,
    '源仓由', @operator_name, '运营。', @contact_details, @nl,
    '二、服务内容', @nl,
    '本平台面向工业品采购方和供货方，提供账号登录、主体认证、产品展示、采购需求发布、供货响应、联系人咨询及会员服务。产品及需求信息用于供需对接，具体商品交易、质量、价格、交付及售后事项由交易双方另行确认。', @nl,
    '三、账号与认证', @nl,
    '你应使用有权使用的手机号登录，并根据所选角色提交真实、准确、合法的认证资料。代表企业操作时，应取得相应授权。不得冒用身份、伪造证照、借用他人账号或向他人泄露登录凭证。资料变化或发现账号异常时，请及时更新资料或联系客服。', @nl,
    '四、信息发布与使用', @nl,
    '你应对产品、采购需求、联系人和附件等所提交内容负责，确保拥有必要权利。不得发布虚假、侵权、违法或与工业供需服务无关的信息。平台可按公示规则审核内容，并对违规内容采取不予展示、下架、限制发布或暂停账号等措施；如有异议，可通过本协议列明的渠道申诉。', @nl,
    '五、会员服务', @nl,
    '需要会员的功能、套餐价格、原价、有效期、开通或续费条件及适用限制，以购买页面和实际订单为准。普通套餐采用单次购买方式，会员到期不会自动扣款。免费套餐的适用条件以页面及服务端校验为准。支付及退款记录用于履约和账务核对，具体会员规则以另行展示并经你同意的会员服务协议为准。', @nl,
    '六、信息安全与个人信息', @nl,
    '平台按照源仓隐私政策处理个人信息。请勿在公开发布内容中包含身份证照片、身份证号码等无关敏感资料。认证资料与公开展示内容分开管理，联系人信息按平台权限规则提供。', @nl,
    '七、服务变更、终止及争议处理', @nl,
    '如服务需要调整，平台将通过适当方式告知。你可以通过客服提出账号注销、资料更正、投诉、会员及退款等请求。服务终止或账号注销后的个人信息处理，按照隐私政策和适用法律执行。争议可先通过客服协商，并依法选择投诉或争议解决途径。', @nl,
    '八、协议更新', @nl,
    '平台将保留协议历史版本。涉及服务范围、收费方式、个人信息处理等实质变化时，将以明显方式通知，并在依法需要时重新取得同意。你可在小程序协议入口查看当前生效版本。'
);

SET @privacy_content = CONCAT(
    '源仓隐私政策', @nl, '生效日期：', @effective_date, @nl,
    '一、运营主体与联系渠道', @nl,
    '个人信息处理者为', @operator_name, '。', @contact_details, @nl,
    '二、手机号登录', @nl,
    '当你主动选择微信手机号登录并同意授权时，我们通过微信提供的能力处理手机号、微信登录标识及登录状态，用于创建账号、身份识别、登录和账号安全。拒绝授权时，相关登录功能无法完成，你仍可使用无需登录的页面。', @nl,
    '三、角色选择与主体认证', @nl,
    '为核验采购方或供货方身份，我们处理你主动提交的角色、姓名、企业名称、营业执照及识别结果、联系人、电话和地址等资料。只有在你选择个人认证并主动提交时，才处理对应的姓名、身份证号码及身份证照片。身份证资料属于敏感个人信息，仅用于所申请认证的真实性核验和必要的争议处理；我们会单独说明必要性及可能影响并取得单独同意。拒绝提交或撤回相关同意可能使该认证无法完成，不影响不依赖认证的功能。', @nl,
    '四、图片和地图权限', @nl,
    '你主动上传产品图片、采购需求附件或认证资料时，会使用相册或相机能力处理你选择或拍摄的图片。你主动选择办公或工厂地址时，会使用地图选择位置能力处理所选地点名称、地址及返回的位置信息。相关权限由你按微信提示授权；拒绝后对应上传或地址选择操作无法完成。我们不以该地址选择功能持续跟踪你的行踪。', @nl,
    '五、供需展示与联系人', @nl,
    '为展示产品和采购需求、匹配供需及联系沟通，我们处理你提交的企业名称、产品资料、需求资料、图片附件、联系人及联系方式。公开发布内容及联系人信息按页面提示和会员、角色等权限规则向其他用户提供。身份证及营业执照认证原件不作为公开产品或需求内容展示。请勿在公开内容中填写无关个人信息。', @nl,
    '六、会员、支付与退款', @nl,
    '你购买会员时，我们处理用户标识、套餐和协议版本、订单号、金额、支付状态、会员期限及必要的支付或退款凭证，用于提供权益、核对账务、处理退款和争议。付款由微信支付完成，我们不获取你的银行卡密码或微信支付密码。普通会员套餐不会到期自动扣款。', @nl,
    '为统计产品和采购需求的访问次数，我们记录用户标识、内容类型、内容标识和点击时间，并向平台运营人员提供汇总报表。必要的服务日志可能记录请求时间、接口路径、状态码、IP 地址和客户端信息，用于运行维护、排查故障和安全保护。', @nl,
    '七、服务提供方', @nl, @service_providers, @nl,
    '我们按实际业务需要与服务提供方约定数据处理范围和保护措施，并依法履行相应义务。涉及向其他个人信息处理者提供个人信息、改变处理目的或需要单独同意的情形时，会另行告知并依法取得同意。', @nl,
    '八、存储、安全与保存期限', @nl,
    '个人信息存储地域：', @storage_region, '。我们采取 HTTPS 传输、访问控制、认证资料私有存储和临时访问链接等措施保护资料。', @nl,
    '账号及发布内容：', @account_retention, '。', @nl,
    '认证资料：', @verification_retention, '。', @nl,
    '订单、支付、退款及必要审计记录：', @order_retention, '。', @nl,
    '产品需求点击记录及必要服务日志：', @analytics_log_retention, '。', @nl,
    '我们按实现处理目的所需的最短时间保存资料；期限届满、处理目的已实现或依法需要删除时，会删除或匿名化。法律规定必须继续保存的资料，将限制为必要的存储和保护处理。', @nl,
    '九、你的权利与联系处理', @nl,
    '你可以通过相关页面或上述客服渠道提出查阅、复制、更正、删除资料、撤回授权、注销账号及投诉等请求。微信权限可在微信相应设置中调整。我们将', @request_response_time,
    '。撤回同意不影响此前基于同意已进行处理的效力；不再满足条件的服务可能无法继续使用，依法必须保存的记录按规定处理。', @nl,
    '十、适用用户与政策变更', @nl,
    '本服务面向工业供需经营活动的用户。请勿冒用他人身份或提交无权处理的未成年人资料。政策更新将保留历史版本，对处理目的、信息种类或对外提供等重要变化会明显通知，并在依法需要时重新取得同意。微信平台展示的用户隐私保护指引应与本政策和实际功能保持一致。'
);

SET @service_content = CONCAT(
    '源仓会员服务协议', @nl, '生效日期：', @effective_date, @nl,
    '一、服务提供者与权益', @nl,
    '会员服务由', @operator_name, '提供。会员可在完成对应账号及主体认证的前提下，使用页面公示的需求发布、供货响应、产品联系人咨询及其他会员功能。具体权益、适用角色及限制以开通时页面和订单为准，会员开通不代表认证自动通过。', @nl,
    '二、套餐价格与有效期', @nl,
    '套餐名称、类型、原价、实际支付金额、服务天数及适用条件在购买前展示，订单固化对应套餐内容及协议版本。普通套餐按次购买，会员到期不会自动扣款。有限期普通会员符合续费条件时，续费时长接续当前剩余有效期。', @nl,
    '三、免费期与购买限制', @nl,
    '0 元套餐在满足页面所列适用条件并完成订单确认后直接开通，无需发起微信支付；已有有效会员不得使用 0 元套餐续费。长期有效会员或其他不满足续费条件的会员，按页面及服务端规则限制重复开通。', @nl,
    '四、支付、开通与异常订单', @nl,
    '收费套餐通过微信支付完成。会员权益以服务端核实的支付结果为准，客户端支付提示不能单独作为到账依据。支付结果暂未确认时，可在会员中心查询或继续处理原订单；请保留订单号以便核对。', @nl,
    '五、退款与权益处理', @nl,
    '如需退款，请通过平台客服提交订单号和申请原因，平台将结合实际履约情况及适用规则处理。当前系统支持整单退款，退款成功后回收该订单尚未使用的会员时长，并保留其他订单购买的权益。退款状态及到账情况以微信支付和平台核实结果为准，法定权利不受本协议限制。', @nl,
    '六、自动续费状态', @nl,
    '当前未开放连续包月签约及自动扣款服务。普通会员无需签署自动续费协议，也不会因服务到期而自动续费。将来如开放自动续费，会另行展示服务规则、扣费安排、取消方式，并在你主动签约后提供。', @nl,
    '七、客服与协议更新', @nl,
    @contact_details, '协议更新采用新增版本方式，已生成订单继续保留其购买时记录的协议号和版本，不因后续编辑而被覆盖。'
);
SET @auto_renew_content = CONCAT('源仓自动续费服务说明', @nl, '生效日期：', @effective_date, @nl,
    '连续包月签约及自动扣款服务当前未开放。普通套餐按次购买，到期不会自动扣款。未来如开放自动续费，将另行提供扣费周期、金额、取消方式等规则，并在用户主动同意和完成签约后提供服务。');

START TRANSACTION;

-- 与后台 UpdateAgreement 使用同一稳定记录锁，串行分配版本号。
SET @agreement_anchor = NULL;
SELECT id INTO @agreement_anchor
FROM vip_agreements ORDER BY id ASC LIMIT 1 FOR UPDATE;

SELECT COALESCE(MAX(version), 0) INTO @user_version
FROM vip_agreements WHERE agreement_key = 'user' FOR UPDATE;
SELECT COALESCE(MAX(version), 0) INTO @privacy_version
FROM vip_agreements WHERE agreement_key = 'privacy' FOR UPDATE;
SELECT COALESCE(MAX(version), 0) INTO @service_version
FROM vip_agreements WHERE agreement_key = 'service' FOR UPDATE;
SELECT COALESCE(MAX(version), 0) INTO @auto_renew_version
FROM vip_agreements WHERE agreement_key = 'auto-renew' FOR UPDATE;

INSERT INTO vip_agreements (agreement_key, title, content, status, version, updated_by)
SELECT 'user', '源仓用户协议', @user_content, 1, @user_version + 1, NULL
WHERE @release_inputs_ready = 1 AND @agreement_anchor IS NOT NULL;
INSERT INTO vip_agreements (agreement_key, title, content, status, version, updated_by)
SELECT 'privacy', '源仓隐私政策', @privacy_content, 1, @privacy_version + 1, NULL
WHERE @release_inputs_ready = 1 AND @agreement_anchor IS NOT NULL;
INSERT INTO vip_agreements (agreement_key, title, content, status, version, updated_by)
SELECT 'service', '源仓会员服务协议', @service_content, 1, @service_version + 1, NULL
WHERE @release_inputs_ready = 1 AND @agreement_anchor IS NOT NULL;
-- 新增停用版本，不覆盖既有自动续费协议；公共接口不会回退到旧的启用版本。
INSERT INTO vip_agreements (agreement_key, title, content, status, version, updated_by)
SELECT 'auto-renew', '自动续费服务说明（暂未开放）', @auto_renew_content, 0, @auto_renew_version + 1, NULL
WHERE @release_inputs_ready = 1 AND @agreement_anchor IS NOT NULL;

-- 用同一批已核对的真实资料替换 contact 页的样例联系方式；空的可选项会停止展示。
UPDATE app_page_contents
SET contact_phone = NULLIF(@contact_phone, ''),
    contact_email = @contact_email,
    address = NULLIF(@operator_address, ''),
    service_hours = NULLIF(@service_hours, ''),
    wechat_account = NULLIF(@wechat_account, ''),
    updated_by = NULL
WHERE page_key = 'contact'
  AND @release_inputs_ready = 1 AND @agreement_anchor IS NOT NULL;

COMMIT;

SELECT CASE
    WHEN COALESCE(@release_inputs_ready, 0) <> 1 THEN 'NOT_PUBLISHED: 请先填写全部 __...__ 变量'
    WHEN @agreement_anchor IS NULL THEN 'NOT_PUBLISHED: 缺少初始协议记录，请先完成数据库迁移'
    ELSE 'PUBLISHED: 已追加新版本，旧版本及订单引用已保留'
END AS publication_result;

SELECT a.id, a.agreement_key, a.title, a.version, a.status, a.updated_at
FROM vip_agreements a
WHERE a.agreement_key IN ('user', 'privacy', 'service', 'auto-renew')
  AND NOT EXISTS (
    SELECT 1 FROM vip_agreements newer
    WHERE newer.agreement_key = a.agreement_key AND newer.version > a.version
  )
ORDER BY a.agreement_key;
