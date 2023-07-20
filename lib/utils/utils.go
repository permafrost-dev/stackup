package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/process"
)

func KillProcessOnWindows(cmd *exec.Cmd) error {
    kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
    kill.Stderr = os.Stderr
    kill.Stdout = os.Stdout
    return kill.Run()
 }

func WaitForStartOfNextMinute() {
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
}

func AbsoluteFilePath(path string) string {
	path, _ = filepath.Abs(path)

	return path
}

func StringToInt(s string, defaultResult int) int {
	i, err := strconv.Atoi(s)

	if err != nil {
		return defaultResult
	}

	return i
}

func Colorize(text string, color int) string {
	// Use the "\x1b[38;5;#m" escape sequence for foreground colors
	return fmt.Sprintf("\x1b[38;5;%dm%s\x1b[0m", color, text)
}

func LoadEnv(filenames ...string) bool {
	// if len(filenames) == 0 {
	// 	filenames = append(filenames, WorkingDir(".env"))
	// }

	err := godotenv.Load()

	if err != nil {
		log.Println("Error loading .env file:", err)
		return false
	}

	return true
}

func RunCommandInPath(input string, dir string, silent bool) *exec.Cmd {
	// Split the input into command and arguments
	parts := strings.Split(input, " ")
	cmd := parts[0]
	// The rest of the parts are the arguments
	args := parts[1:]

	c := exec.Command(cmd, args...)
	if !silent {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}
	c.Dir = dir

	if err := c.Run(); err != nil {
		log.Fatal(err)
	}

	return c
}

func RunCommand(input string) *exec.Cmd {
	// Split the input into command and arguments
	parts := strings.Split(input, " ")
	cmd := parts[0]
	// The rest of the parts are the arguments
	args := parts[1:]

	c := exec.Command(cmd, args...)

	// Connect the command's output to the standard output
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	// Start the command
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}

	return c
}

func RunCommandSilent(input string) *exec.Cmd {
	// Split the input into command and arguments
	parts := strings.Split(input, " ")
	cmd := parts[0]
	// The rest of the parts are the arguments
	args := parts[1:]

	c := exec.Command(cmd, args...)

	c.Stdout = nil
	c.Stderr = nil

	// Start the command
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}

	return c
}

func RunCommandEx(input string, cwd string) *exec.Cmd {
	// Split the input into command and arguments
	parts := strings.Split(input, " ")
	cmd := parts[0]
	// The rest of the parts are the arguments
	args := parts[1:]

	c := exec.Command(cmd, args...)
	c.Dir = cwd

	// Connect the command's output to the standard output
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	// Start the command
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}

	return c
}

func StartCommand(input string) (*exec.Cmd, error) {
	// Split the input into command and arguments
	parts := strings.Split(input, " ")
	cmd := parts[0]
	args := parts[1:]

	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c, nil
}

func WorkingDir(filenames ...string) string {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}

	parts := make([]string, 0)
	parts = append(parts, dir)
	parts = append(parts, filenames...)

	dir = path.Join(parts...)

	return dir
}

type InitializableStruct struct {
	Init *func(c interface{})
}

type InitFunc = func(c interface{})

func MatchDotProperties(s string) []string {
	// Regular expression to match dot properties
	r := regexp.MustCompile(`\$\{(\w+\.\w+)*\}`)

	return r.FindAllString(s, -1)
}

func GetDotProperty(obj interface{}, property string) interface{} {
	props := strings.Split(property, ".")
	value := reflect.ValueOf(obj)

	for _, prop := range props {
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if value.Kind() != reflect.Struct {
			fmt.Println("Not a struct")
			return nil
		}

		value = value.FieldByName(prop)

		if !value.IsValid() {
			fmt.Println("Property not found")
			return nil
		}
	}

	return value
}

func GetDotPropertyStruct(obj interface{}, property string) interface{} {
	props := strings.Split(property, ".")
	value := reflect.ValueOf(obj)

	for _, prop := range props {
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		if value.Kind() != reflect.Struct {
			fmt.Println("Not a struct")
			return nil
		}

		value = value.FieldByName(prop)

		if !value.IsValid() {
			fmt.Println("Property not found")
			return nil
		}
	}

	// Check if the value implements Initializable interface
	// var initializable lib.Initializable

	if value.CanInterface() {
		return value.Interface()
	}

	return nil

	// if value.CanInterface() && value.Type().Implements(reflect.TypeOf(&initializable).Elem()) {
	// 	return value.Interface().(lib.Initializable)
	// } else {
	// 	fmt.Println("Value does not implement Initializable interface")
	// 	return nil
	// }
}

func CallMethod(i interface{}, methodName string, args ...interface{}) {
	value := reflect.ValueOf(i)

	// Convert the arguments to reflect.Value
	rargs := make([]reflect.Value, len(args))
	for i, arg := range args {
		if reflect.ValueOf(arg).Kind().String() != "zero" {
			rargs[i] = reflect.ValueOf(arg)
		}
	}

	// Call the method
	// var result []reflect.Value

	fmt.Printf("%s, %v", methodName, value.MethodByName("Init"))

	// // Print the results
	// for _, res := range result {
	// 	fmt.Println("Result:", res)
	// }
}

func ReplaceConfigurationKeyVariablesInStruct(str string, inputStruct interface{}, parentKey string) string {
	if strings.Contains(str, "${") {
		return GetDotProperty(inputStruct, strings.Trim(str, "${}")).(string)
	}

	val := reflect.Indirect(reflect.ValueOf(inputStruct))
	t := val.Type()

	// fmt.Printf("====>>> %v", val) //Sreflect.MapOf(t, val.Type()))

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := val.Field(i)
		// if value.Kind() == reflect.Struct {
		// 	str = ReplaceConfigurationKeyVariablesInStruct(str, value.Interface(), parentKey+"."+field.Name)
		// } else {
		variable := "${" + parentKey + "." + field.Name + "}"
		str = strings.ReplaceAll(str, variable, fmt.Sprint(value.Interface()))
		// }
	}

	return str
}

func ReplaceConfigurationKeyVariablesInMap(str string, replacements interface{}, parentKey string) string {
	fmt.Printf("%v", GetStructKeys(replacements))
	for _, value := range GetStructKeys(replacements) {

		// if parentKey != "" {
		newKey := parentKey + "." + value
		// } else {
		// 	newKey = strconv.FormatInt(int64(key), 10)
		// }

		// fmt.Println("**********newkey=%v", newKey)

		// if castedValue, ok := value; ok {
		//str = ReplaceConfigurationKeyVariablesInMap(str, value, newKey)
		// } else {
		variable := "${" + newKey + "}"
		str = strings.ReplaceAll(str, variable, fmt.Sprint(value))
		// }
	}

	return str
}

func GetStructKeys(data interface{}) []string {
	val := reflect.ValueOf(data)
	typ := val.Type()

	// Make sure the provided data is a struct
	if typ.Kind() != reflect.Struct {
		return nil
	}

	var keys []string

	// Iterate over the fields of the struct
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		keys = append(keys, field.Name)
	}

	return keys
}

func GetStructSubKeys(data interface{}, subKey string) []string {
	val := reflect.ValueOf(data)
	typ := val.Type()

	// Make sure the provided data is a struct
	if typ.Kind() != reflect.Struct {
		return nil
	}

	var keys []string

	// Iterate over the fields of the struct
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.Name == subKey {
			return GetStructKeys(typ.Field(i))
		}
		keys = append(keys, field.Name)
	}

	return keys
}

// func GetStructSubKeys(obj interface{}, data interface{}) []string {
//     keys := GetStructKeys(obj)

// for _, key := range keys {

// }
// 	val := reflect.ValueOf(data)
// 	typ := val.Type()

// 	// Make sure the provided data is a struct
// 	if typ.Kind() != reflect.Struct {
// 		return nil
// 	}

// 	var keys []string

// 	// Iterate over the fields of the struct
// 	for i := 0; i < val.NumField(); i++ {
// 		field := typ.Field(i)
// 		keys = append(keys, field.Name)
// 	}

// 	return keys
// }

func SetStructKeyValue(data interface{}, key string, value interface{}) error {
	val := reflect.ValueOf(data)

	// Make sure the provided data is a pointer to a struct
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("data should be a pointer to a struct")
	}

	fieldValue := val.Elem().FieldByName(key)
	if !fieldValue.IsValid() {
		return fmt.Errorf("key '%s' does not exist in the struct", key)
	}

	// Make sure the provided value is assignable to the struct field
	if !reflect.TypeOf(value).AssignableTo(fieldValue.Type()) {
		return fmt.Errorf("value type mismatch for key '%s'", key)
	}

	fieldValue.Set(reflect.ValueOf(value))

	return nil
}

func GetStructKeyValue(data interface{}, key string) reflect.Value {
	val := reflect.ValueOf(data)

	// Make sure the provided data is a pointer to a struct
	// if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
	// 	return
	// }

	return val.Elem().FieldByName(key)
}

func ParseDurationString(duration string) int64 {
	// Parse the duration string
	d, err := time.ParseDuration(duration)
	if err != nil {
		return 0
	}

	// Convert the duration to seconds and return it as an int64
	return int64(d.Milliseconds())
}

func CheckProcessCpuLoad(pid int32) float64 {
	p, err := process.NewProcess(pid)
	if err != nil {
		return 0.0
	}

	result, _ := p.CPUPercent()
	return result
}

func CheckProcessMemoryUsage(pid int32) float64 {
	p, err := process.NewProcess(pid)

	if err != nil {
		return 0.0
	}

	result, _ := p.MemoryPercent()

	return float64(result)
}

func FindFirstExistingFile(filenames []string) (string, error) {
	for _, filename := range filenames {
		if _, err := os.Stat(filename); err == nil {
			return filename, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	return "", fmt.Errorf("not found")
}
