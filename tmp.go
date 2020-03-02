package main
import(
	"fmt"
	"strings"
)
func replaceNth(s string, n int) string {
	//"enabled":true
	old := "mao:true"
	new := "mao:false"
	i := 0
	for m := 1; m <= n; m++ {
		x := strings.Index(s[i:], old)
		if x < 0 {
			break
		}
		i += x
		if m == n {
			return s[:i] + new + s[i+len(old):]
		}
		i += len(old)
	}
	return s
}
func main(){
	s:="mao:true,mao:true,mao:true"
	s = replaceNth(s,3)
	s = replaceNth(s,2)
	s = replaceNth(s,1)
	fmt.Println(s)
}