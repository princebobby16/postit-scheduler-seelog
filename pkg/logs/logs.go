package logs

import (
	see "github.com/cihub/seelog"
	"log"
)

var Logger see.LoggerInterface

func init() {
	Logger = see.Disabled
	loadAppConfig()
}

func loadAppConfig () {

	appConfig := `
<seelog type="sync">
	<outputs>
		<filter levels="trace">
			<console formatid="plain"/>
			<file path="./pkg/logs/common.log"/>
		</filter>

		<filter levels="info">
			<console formatid="plain"/>
			<file path="./pkg/logs/common.log"/>
		</filter>

		<filter levels="warn">
			<console formatid="plain" />
			<file path="./pkg/logs/common.log"/>
		</filter>

		<filter levels="error">
			<console formatid="error"/>
			<file path="./pkg/logs/error.log"/>
			<smtp formatid="erroremail" senderaddress="shiftrgh@gmail.com" sendername="Scheduler Microservice" hostname="smtp
				.gmail.com" hostport="587" username="shiftrgh@gmail.com" password="yoforreal.com">
				<recipient address="shiftrgh@gmail.com"/>
			</smtp>
		</filter>
		
		<filter levels="critical">
			<console formatid="critical"/>
			<file path="./pkg/logs/critical.log"/>
			<smtp formatid="criticalemail" 
				senderaddress="shiftrgh@gmail.com" 
				sendername="Scheduler Microservice" 
				hostname="smtp.gmail.com" 
				hostport="587" 
				username="shiftrgh@gmail.com" 
				password="yoforreal.com">
				<recipient address="shiftrgh@gmail.com"/>
			</smtp>
		</filter>
		
	</outputs>
	<formats>
		<format id="plain" format="%Date/%Time [%LEVEL] %Msg%n" />
		<format id="error" format="%Date/%Time [%LEVEL] %RelFile %Func %Line %Msg%n" />
		<format id="critical" format="%Date/%Time [%LEVEL] %RelFile %Func %Line %Msg%n" />
		<format id="erroremail" format="Minor error on our server! %n%Time %Date [%LEVEL] %FullPath %n%RelFile %n%File  %n%Func %n%Msg%n %nSent by PostIt Scheduler Micro-Service"/>
		<format id="criticalemail" format="Critical error on our server! %n%Time %Date [%LEVEL] %FullPath %n%RelFile %n%File  %n%Func %n%Msg%n %nSent by PostIt Scheduler Micro-Service"/>
	</formats>
</seelog>
`

	logger, err := see.LoggerFromConfigAsBytes([]byte(appConfig))
	if err != nil {
		log.Println(err)
		return
	}
	//err = see.ReplaceLogger(logger)
	//if err != nil {
	//	panic(err)
	//	return
	//}

	UseLog(logger)
}

func UseLog(log see.LoggerInterface)  {
	Logger = log
}
