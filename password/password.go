package password

import (
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"

	"github.com/adamesong/go-util/random"
	"golang.org/x/crypto/pbkdf2"
)

// MakePassword 用于将明文密码转为加密后的密码。这里没有判断密码是否为空。
// django 密码加密算法的 go 语言版本。
// Django 的实现细节请参考 Python 和 django 文档：
// django.contrib.auth.hashers.make_password
// django.utils.crypto import pbkdf2
// hashlib.sha256
// base64
// 参考：https://studygolang.com/articles/5262
func Encrypt(password, saltString string, iterations int) string {
	pwd := []byte(password)
	var salt []byte
	if saltString == "" {
		salt = []byte(random.RandomString(12)) // 盐，是一个随机字符串，每一个用户都不一样，在这里我们随机选择 12个字符的字符串 作为盐
	} else {
		salt = []byte(saltString)
	}
	if iterations == 0 {
		iterations = 120000 // 加密算法的迭代次数，120000 次
	}
	digest := sha256.New // digest 算法，使用 sha256

	// 第一步：使用 pbkdf2 算法加密
	dk := pbkdf2.Key(pwd, salt, iterations, 32, digest)

	// 第二步：Base64 编码
	str := base64.StdEncoding.EncodeToString(dk)

	// 第三步：组合加密算法、迭代次数、盐、密码和分割符号 "$"
	return "pbkdf2_sha256" + "$" + strconv.FormatInt(int64(iterations), 10) + "$" + string(salt) + "$" + str
}

// IsSamePassword用来判断password字符串encode后是否与encoded一致
// pbkdf2_sha256$120000$ONRhfKsUOHoF$xHEtXKw7u4F5hhdEj8sMUwHOcP06KBFliFnYzF7qYnw= 包括了4个部分，分别是：
// pbkdf2_sha256 100000 M1BIGL7NBnF1 gACxYtYQItPQ73FiWKYnYbCDdJeV2zlhobVcdkTd/Lg=
func IsSame(password, encoded string) bool {
	encodedArray := strings.Split(encoded, "$")
	iterations, _ := strconv.Atoi(encodedArray[1])
	salt := encodedArray[2]
	toBeVerified := Encrypt(password, salt, iterations)
	return toBeVerified == encoded
}

// todo 验证password的设置规则，如
// if len(password) < self.min_length:
// DEFAULT_USER_ATTRIBUTES = ('username', 'first_name', 'last_name', 'email')
//
//    def __init__(self, user_attributes=DEFAULT_USER_ATTRIBUTES, max_similarity=0.7):
//        self.user_attributes = user_attributes
//        self.max_similarity = max_similarity
//
//    def validate(self, password, user=None):
//        if not user:
//            return
//
//        for attribute_name in self.user_attributes:
//            value = getattr(user, attribute_name, None)
//            if not value or not isinstance(value, str):
//                continue
//            value_parts = re.split(r'\W+', value) + [value]
//            for value_part in value_parts:
//                if SequenceMatcher(a=password.lower(), b=value_part.lower()).quick_ratio() >= self.max_similarity:
//                    try:
//                        verbose_name = str(user._meta.get_field(attribute_name).verbose_name)
//                    except FieldDoesNotExist:
//                        verbose_name = attribute_name
//                    raise ValidationError(
//                        _("The password is too similar to the %(verbose_name)s."),
//                        code='password_too_similar',
//                        params={'verbose_name': verbose_name},
