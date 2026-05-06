package generated_program;

import com.pulumi.Pulumi;
import com.pulumi.pulumiservice.v2_insights.Account;
import com.pulumi.pulumiservice.v2_insights.AccountArgs;
import com.pulumi.pulumiservice.v2_insights.ScheduledScanSettings;
import com.pulumi.pulumiservice.v2_insights.ScheduledScanSettingsArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            var config = ctx.config();
            var serviceOrg = config.get("serviceOrg").orElse("service-provider-test-org");
            var accountSuffix = config.get("accountSuffix").orElse("dev");
            var insightsEnv = config.get("insightsEnvironment").orElse("insights/credentials");

            var account = new Account("account",
                AccountArgs.builder()
                    .orgName(serviceOrg)
                    .accountName("v2-insights-" + accountSuffix)
                    .provider("aws")
                    .environment(insightsEnv)
                    .scanSchedule("none")
                    .build());

            new ScheduledScanSettings("scanSettings",
                ScheduledScanSettingsArgs.builder()
                    .orgName(serviceOrg)
                    .accountName(account.accountName())
                    .paused(true)
                    .scheduleCron("0 6 * * *")
                    .build());

            ctx.export("accountName", account.accountName());
        });
    }
}
