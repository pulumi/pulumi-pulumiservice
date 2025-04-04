// *** WARNING: this file was generated by pulumi. ***
// *** Do not edit by hand unless you're certain you know what you are doing! ***

package com.pulumi.pulumiservice;

import com.pulumi.core.Output;
import com.pulumi.core.annotations.Import;
import com.pulumi.exceptions.MissingRequiredPropertyException;
import java.lang.Double;
import java.lang.String;
import java.util.List;
import java.util.Objects;
import java.util.Optional;
import javax.annotation.Nullable;


public final class TeamArgs extends com.pulumi.resources.ResourceArgs {

    public static final TeamArgs Empty = new TeamArgs();

    /**
     * Optional. Team description.
     * 
     */
    @Import(name="description")
    private @Nullable Output<String> description;

    /**
     * @return Optional. Team description.
     * 
     */
    public Optional<Output<String>> description() {
        return Optional.ofNullable(this.description);
    }

    /**
     * Optional. Team display name.
     * 
     */
    @Import(name="displayName")
    private @Nullable Output<String> displayName;

    /**
     * @return Optional. Team display name.
     * 
     */
    public Optional<Output<String>> displayName() {
        return Optional.ofNullable(this.displayName);
    }

    /**
     * The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for &#34;github&#34; teams.
     * 
     */
    @Import(name="githubTeamId")
    private @Nullable Output<Double> githubTeamId;

    /**
     * @return The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for &#34;github&#34; teams.
     * 
     */
    public Optional<Output<Double>> githubTeamId() {
        return Optional.ofNullable(this.githubTeamId);
    }

    /**
     * List of Pulumi Cloud usernames of team members.
     * 
     */
    @Import(name="members")
    private @Nullable Output<List<String>> members;

    /**
     * @return List of Pulumi Cloud usernames of team members.
     * 
     */
    public Optional<Output<List<String>>> members() {
        return Optional.ofNullable(this.members);
    }

    /**
     * The team&#39;s name. Required for &#34;pulumi&#34; teams.
     * 
     */
    @Import(name="name")
    private @Nullable Output<String> name;

    /**
     * @return The team&#39;s name. Required for &#34;pulumi&#34; teams.
     * 
     */
    public Optional<Output<String>> name() {
        return Optional.ofNullable(this.name);
    }

    /**
     * The name of the Pulumi organization the team belongs to.
     * 
     */
    @Import(name="organizationName", required=true)
    private Output<String> organizationName;

    /**
     * @return The name of the Pulumi organization the team belongs to.
     * 
     */
    public Output<String> organizationName() {
        return this.organizationName;
    }

    /**
     * The type of team. Must be either `pulumi` or `github`.
     * 
     */
    @Import(name="teamType", required=true)
    private Output<String> teamType;

    /**
     * @return The type of team. Must be either `pulumi` or `github`.
     * 
     */
    public Output<String> teamType() {
        return this.teamType;
    }

    private TeamArgs() {}

    private TeamArgs(TeamArgs $) {
        this.description = $.description;
        this.displayName = $.displayName;
        this.githubTeamId = $.githubTeamId;
        this.members = $.members;
        this.name = $.name;
        this.organizationName = $.organizationName;
        this.teamType = $.teamType;
    }

    public static Builder builder() {
        return new Builder();
    }
    public static Builder builder(TeamArgs defaults) {
        return new Builder(defaults);
    }

    public static final class Builder {
        private TeamArgs $;

        public Builder() {
            $ = new TeamArgs();
        }

        public Builder(TeamArgs defaults) {
            $ = new TeamArgs(Objects.requireNonNull(defaults));
        }

        /**
         * @param description Optional. Team description.
         * 
         * @return builder
         * 
         */
        public Builder description(@Nullable Output<String> description) {
            $.description = description;
            return this;
        }

        /**
         * @param description Optional. Team description.
         * 
         * @return builder
         * 
         */
        public Builder description(String description) {
            return description(Output.of(description));
        }

        /**
         * @param displayName Optional. Team display name.
         * 
         * @return builder
         * 
         */
        public Builder displayName(@Nullable Output<String> displayName) {
            $.displayName = displayName;
            return this;
        }

        /**
         * @param displayName Optional. Team display name.
         * 
         * @return builder
         * 
         */
        public Builder displayName(String displayName) {
            return displayName(Output.of(displayName));
        }

        /**
         * @param githubTeamId The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for &#34;github&#34; teams.
         * 
         * @return builder
         * 
         */
        public Builder githubTeamId(@Nullable Output<Double> githubTeamId) {
            $.githubTeamId = githubTeamId;
            return this;
        }

        /**
         * @param githubTeamId The GitHub ID of the team to mirror. Must be in the same GitHub organization that the Pulumi org is backed by. Required for &#34;github&#34; teams.
         * 
         * @return builder
         * 
         */
        public Builder githubTeamId(Double githubTeamId) {
            return githubTeamId(Output.of(githubTeamId));
        }

        /**
         * @param members List of Pulumi Cloud usernames of team members.
         * 
         * @return builder
         * 
         */
        public Builder members(@Nullable Output<List<String>> members) {
            $.members = members;
            return this;
        }

        /**
         * @param members List of Pulumi Cloud usernames of team members.
         * 
         * @return builder
         * 
         */
        public Builder members(List<String> members) {
            return members(Output.of(members));
        }

        /**
         * @param members List of Pulumi Cloud usernames of team members.
         * 
         * @return builder
         * 
         */
        public Builder members(String... members) {
            return members(List.of(members));
        }

        /**
         * @param name The team&#39;s name. Required for &#34;pulumi&#34; teams.
         * 
         * @return builder
         * 
         */
        public Builder name(@Nullable Output<String> name) {
            $.name = name;
            return this;
        }

        /**
         * @param name The team&#39;s name. Required for &#34;pulumi&#34; teams.
         * 
         * @return builder
         * 
         */
        public Builder name(String name) {
            return name(Output.of(name));
        }

        /**
         * @param organizationName The name of the Pulumi organization the team belongs to.
         * 
         * @return builder
         * 
         */
        public Builder organizationName(Output<String> organizationName) {
            $.organizationName = organizationName;
            return this;
        }

        /**
         * @param organizationName The name of the Pulumi organization the team belongs to.
         * 
         * @return builder
         * 
         */
        public Builder organizationName(String organizationName) {
            return organizationName(Output.of(organizationName));
        }

        /**
         * @param teamType The type of team. Must be either `pulumi` or `github`.
         * 
         * @return builder
         * 
         */
        public Builder teamType(Output<String> teamType) {
            $.teamType = teamType;
            return this;
        }

        /**
         * @param teamType The type of team. Must be either `pulumi` or `github`.
         * 
         * @return builder
         * 
         */
        public Builder teamType(String teamType) {
            return teamType(Output.of(teamType));
        }

        public TeamArgs build() {
            if ($.organizationName == null) {
                throw new MissingRequiredPropertyException("TeamArgs", "organizationName");
            }
            if ($.teamType == null) {
                throw new MissingRequiredPropertyException("TeamArgs", "teamType");
            }
            return $;
        }
    }

}
