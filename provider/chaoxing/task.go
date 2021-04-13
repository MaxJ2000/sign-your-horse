package chaoxing

import (
	"fmt"
	"net/url"
	"regexp"
	"sign-your-horse/common"
	"strings"
	"time"

	"github.com/imroc/req"
)

func (c *ChaoxingProvider) Task() {
	extractTasksRegex := regexp.MustCompile(`activeDetail\(\d+,`)
	extrackTaskIDRegex := regexp.MustCompile(`\d+`)
	extrackTasksTypeRegex := regexp.MustCompile(`<a href="javascript:;" shape="rect">.*</a>`)
	extrackTaskTypeRegex := regexp.MustCompile(`>(.*)<`)

	r := req.New()
	tasks, err := r.Get(
		fmt.Sprintf("https://mobilelearn.chaoxing.com/widget/pcpick/stu/index?courseId=%s&jclassId=%s",
			c.CourseID,
			c.ClassID),
		req.Header{
			"Cookie":     c.Cookie,
			"User-Agent": c.UserAgent,
		},
	)
	if err != nil {
		c.PushMessageWithAlias("get task list failed: " + err.Error())
	} else {
		taskListString := tasks.String()
		if len(taskListString) == 0 {
			c.PushMessageWithAlias("get task list failed: empty page")
			return
		}
		finishedSepIndex := strings.Index(taskListString, "已结束")
		if finishedSepIndex == -1 {
			c.PushMessageWithAlias("invalid task page, maybe you need login?")
			return
		}
		taskListString = taskListString[:finishedSepIndex]
		tasksString := extractTasksRegex.FindAll([]byte(taskListString), -1)
		tasksTypeString := extrackTasksTypeRegex.FindAll([]byte(taskListString), -1)
		if len(tasksString) == 0 && c.Verbose {
			common.LogWithModule(c.Alias, " no task in list at %s", time.Now().String())
		} else {
			for i, task := range tasksString {
				taskTypeArr := extrackTaskTypeRegex.FindSubmatch(tasksTypeString[i])
				taskType := ""
				if len(taskTypeArr) > 1 {
					taskType = string(taskTypeArr[1])
				}
				taskID := extrackTaskIDRegex.Find(task)
				var request string
				if taskType == "位置签到" && c.Latitude != "" && c.Longitude != "" && c.Address != "" {
					request = fmt.Sprintf("https://mobilelearn.chaoxing.com/pptSign/stuSignajax?name=%s&address=%s&activeId=%s&uid=%s&clientip=&useragent=&latitude=%s&longitude=%s&fid=0&appType=15", url.PathEscape(c.Name), url.PathEscape(c.Address), string(taskID), c.UserID, c.Latitude, c.Longitude)
				} else {
					request = fmt.Sprintf("https://mobilelearn.chaoxing.com/pptSign/stuSignajax?name=&activeId=%s&uid=%s&clientip=&useragent=&latitude=-1&longitude=-1&fid=0&appType=15", string(taskID), c.UserID)
				}
				time.Sleep(time.Second * time.Duration(c.TaskInterval))
				resp, err := r.Get(
					request,
					req.Header{
						"Cookie":     c.Cookie,
						"User-Agent": c.UserAgent,
					},
				)
				if err != nil {
					c.PushMessageWithAlias("task " + taskType + " " + string(taskID) + " sign in failed: " + err.Error())
				} else {
					c.PushMessageWithAlias("task " + taskType + " " + string(taskID) + " sign in result: " + resp.String())
				}
			}
		}
	}
}

func (c *ChaoxingProvider) PushMessageWithAlias(msg string) error {
	return c.PushMessageCallback(c.Alias, msg)
}
