# Test

1. 测试文件命名：后缀为_test.go，一般与被测文件放在一起。文件名不必一定和被测试文件的名称一样，但最好保持一致。

2. TestMain()为包内公共函数，统一单独放至main_test.go中，与其他用例保持隔离性。

   e.g.

   <img src="./images/main_test.png"/>

3. 功能测试函数命名规范：Test+FuncName

   e.g.

   ```go
   func TestMyFunction(t *testing.T) {
       // 测试代码
   }
   ```

4. 功能测试函数内容规范：

   ```go
   func Test+FuncName(t *testing.T) {
       Convey("func {被测函数名} {预期结果 succeeded}", t, test+被测函数内容)
   	Convey("func {被测函数名} {预期结果 failed, + 错误原因}", t, test+被测函数内容+Err+简述错误原因)
   }
   ```

   e.g.

   <img src="./images/convey_standard.png"/>

5. 测试用例代码编写顺序规范：

   Part1：测试前置条件设置；

   Part2：打桩；

   Part3：调用被测对象；

   Part4：断言。

   e.g.

   <img src="./images/case_order.png"/>

6. 多个用例共用一个初始化/打桩函数，可以通过一个功能测试函数中进行Convey嵌套的方式解决。若无依赖关系，并列即可。测试结果也会分层级显示。
   
   Convey嵌套原则：

   每执行一次最内层的Convey都会从最外层开始逐层执行Convey的，且会略过内层已经执行过的Convey。

   e.g. 

   ```go
   func TestConvey(t *testing.T) {
       Convey("nesting convey", t, func() {
           fmt.Println("loop...")
           Convey("first", func() {
               fmt.Println("一")
               Convey("first-1", func() {
                   fmt.Println("1")
               })
               Convey("first-2", func() {
                   fmt.Println("2")
               })
           })
           Convey("second", func() {
               fmt.Println("二")
               Convey("second-1", func() {
                   fmt.Println("1")
               })
               Convey("second-2", func() {
                   fmt.Println("2")
               })
           })
       })
   }
   
   // 执行结果：
   loop...
   一
   1
   loop...
   一
   2
   loop...
   二
   1
   loop...
   二
   2
   ```

7. 打桩函数的error，可用本包中的ErrTest变量：

    ```go
    var ErrTest = errors.New("test error")
    ```

8. 本包将ut代码中的公共函数抽取出来，配合框架使用即可简化ut代码。

   若包初始化只需进行日志初始化：

   ```go
   func TestMain(m *testing.M) {
   	patches := gomonkey.ApplyFuncReturn(func1, func2)
   	tcBase := &test.TcBase{}
   	test.RunWithPatches(tcBase, m, patches)
   }
   ```

   若需要进行数据库与表的创建：

   ```go
   func TestMain(m *testing.M) {
   	tables := make([]interface{}, 0)
   	tcBaseWithDb := &test.TcBaseWithDb{
           DbPath: "xxx",
   		Tables: append(tables, &table1{}, &table1{}),
   	}
   	patches := gomonkey.ApplyFunc(func1, func2)
   	test.RunWithPatches(tcBaseWithDb, m, patches)
   }
   ```

   若需要定制化初始化流程，重新实现tcModule接口即可：

   ```go
   // TcXXX struct for test case base
   type TcXXX struct{}
    
   // Setup pre-processing
   func (tc *TcXXX) Setup() error {
   	// 自定义setup
   	return nil
   }
    
   // Teardown post-processing
   func (tc *TcXXX) Teardown() {
   	// 自定义teardown
   }
   ```

9. 被测函数相对简单时，相似的用例写在一个功能测试函数中比较清晰，但是会有函数深度过大和功能不单一的问题（mr中提屏蔽即可）。

   ```go
   func TestSet(t *testing.T) {
       Convey("func Xxx succeeded", t, func() {
           ...
       })
    
       convey.Convey("func Xxx failed, reason1", t, func() {
           ...
       })
    
       convey.Convey("func Xxx failed, reason2", t, func() {
           ...
       })
    
       convey.Convey("func Xxx failed, reason3", t, func() {
           ...
       })
   }
   ```

10. 导入的Convey包，可以使用别名.代替，简化书写。

   ```go
   package xxx
   
   import (
       "testing"
   
       . "github.com/smartystreets/goconvey/convey"
   )
   
   func TestXxx(t *testing.T) {
       Convey("func MyFunction success", t, func() {
           ...
           So(err, ShouldBeNil)
       })
   }
   ```

11. 在setup或生成一些用例中需要用到的数据时，出现error不可直接return，需要向上抛异常或者进行error断言。否则，会导致用例被跳过。

   原因：

   ① 在setup的阶段，出现了error会终止流程，后面所有的测试用例将不会再执行；

   ② 在用例的准备阶段，生成用例中需要用到的数据过程时，如果出现了error直接返回，后面的用例同样不会再执行。因为此时捕捉到的退出码为0，表示正常退出，不会报错。