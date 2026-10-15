# Complete the incomplete speccheck bindings

## Scope, starting point, and approved decisions

- Work in `/projects/elasticsearch`, module `github.com/disaster37/elasticsearch/v9`, on existing branch `feat/auto-retry-multi-address` (PR #1, base v9). Inspected HEAD: `cc9ce3aeb1d8834ebecbe2b06898e948bedbf646`.
- Pinned Elasticsearch v9.0.0 spec: `/tmp/es-spec/rest-api-spec/src/main/resources/rest-api-spec/api`; inspected clone HEAD `112859b85d50de2a7e63f73c8fc70b99eea24291`.
- Starting audit: checked=512, implemented=158, known-missing=354, ignored-params=21. CI green is user-reported; this planning session had file tools only and ran no commands/tests.
- Target **175 endpoints**: **173 incomplete-binding reasons + 2 text-structure route reasons**. Expected final audit: **checked=512, implemented=333, known-missing=179, ignored-params=25**. Counts are acceptance expectations, never a reason to invent coverage.
- Residual baseline: **45 no-service + 134 no-method** entries, including **48 experimental and 131 stable** endpoints. Experimental metadata is orthogonal to absence of implementation. Complete the five experimental incomplete-binding candidates too: cat.tasks, features.reset_features, search_application.list/search, searchable_snapshots.clear_cache. Stability is not a coverage filter.
- No implementation of absent services/methods, generic inference provider aliases, new client accessors, response-schema parity, every HTTP verb, `_common` auditing, spec-pin upgrades, dependencies, CI policy changes, or unrelated refactoring.
- Preserve all original service interfaces/signatures. Add optional fields directly where safe; missing carriers use opt-in extensions. New extension signatures with 3+ inputs after ctx use `WithRequest`; simpler ones use `WithParams`.
- User permits direct correction of proven legacy route/verb defects outside text_structure. For text_structure specifically, preserve the two legacy methods and add correctly routed extensions. No data migration; advertise changed legacy routes as bug fixes and extensions as opt-in additions.
- Shared timeout extensions use three exact carriers, not a universal carrier. Keep CommonParams unchanged: it is a named field/alias to types.CommonParams, not a mechanism for endpoint-local coverage.
- mtermvectors gets a dedicated extension carrier, not Ids on the TermVectorsParams alias. ForceMerge gets an extension response retaining both task and synchronous shard results. Keep SecurityGetAPIKeyParams.ID and legacy WithLimitedPrivileges wire behavior; add WithLimitedBy separately.

## 1. Freeze a deterministic target manifest before editing

Use standard-library Go AST parsing (`go/parser`, `go/ast`, `strconv.Unquote`) and existing `generator.ScanAPI`, `APIModel.Aliases/Methods/AccessibleFields`, spec loader, and goccy/go-json. Do not add a framework or infer registrations from names. A temporary analyzer under `/tmp` is sufficient; do not add automatic baseline generation to the repo.

1. Locate only the `knownMissingEndpoints` map literal in baseline.go. Extract each literal endpoint key, Reason, Status, Stability, Visibility, Tracking. Reject duplicate keys, missing values, and unrecognized reason classifications.
2. Match the full unquoted Reason against:
   `^([A-Za-z][A-Za-z0-9]*Service)\.([A-Za-z][A-Za-z0-9]*) exists but binding is incomplete \((.+)\)$`.
   Split the third capture on `; `; permit exactly one `missing params: ` clause and/or one `unmapped parts: ` clause. Split clause values on comma, trim, reject empty/duplicate names, sort. Preserve the original reason verbatim alongside parsed values.
3. Route candidates are an explicit allowlist of the two exact existing reason strings at baseline.go's text_structure.find_field_structure/find_message_structure entries. Assert both are present. Do not select arbitrary reasons by substring.
4. Exclude full matches of `^no service in client for namespace .*` and `^no [A-Za-z][A-Za-z0-9]*Service\.[A-Za-z][A-Za-z0-9]* method$`; preserve these entries untouched. A new unmatched reason is a stop/review condition.
5. Join names to `go run ./internal/speccheck -spec "$SPEC_DIR" -inventory` JSON rows for exact spec params/parts/metadata, then resolve the **actual** interface method and concrete transport. Inventory is heuristic: unregistered rows currently leave ParamsStruct and gap arrays empty; it cannot supply the carrier matrix on its own.
6. Baseline method spellings are not authoritative: e.g. Forcemerge -> ForceMerge; GetApiKey/CreateApiKey -> GetAPIKey/CreateAPIKey. Use explicit overrides verified against interfaces, never broad case-insensitive matching.
7. Resolve a method's params argument or request.Params recursively through local aliases; canonicalize carrier identities (CreateParams -> IndexParams; CloseParams -> OpenParams; SplitParams/CloneParams -> ShrinkParams; SimulateTemplateParams -> SimulateIndexTemplateParams). Retain original qualified carrier names for bindings. Separate candidates with no carrier, and record planned shared extension carrier.
8. Emit sorted manifest rows: endpoint, namespace (`root` for undotted names), original reason, actual service/method, canonical/shared carrier, source file, ParamsSource, QueryArgs, PathSources, missing field/type/wire triples, transport test name, batch, disposition. Group by namespace and canonical carrier. Sort shared-carrier groups by number of endpoints fully unlocked, descending, then carrier name; route/response safety overrides cheap-yield ordering.
9. Assert 354 baseline entries partition into 173 binding + 2 route + 45 no-service + 134 no-method, and manifest names match exactly the batch lists below. Freeze the original manifest outside source so subsequent baseline removals do not lose the target set. Put its reviewed matrix/results in internal/speccheck/README.md as batches land.

Namespace target counts: async_search 1; cat 4; ccr 13; cluster 15; dangling_indices 2; enrich 5; eql 2; features 2; ilm 7; indices 38; inference 1; ingest 3; license 5; ml 4; nodes 1; root 16; search_application 2; searchable_snapshots 2; security 15; slm 9; snapshot 12; synonyms 3; text_structure 2; transform 3; watcher 6; xpack 2. Total 175.

## 2. Fixed field, transport, and file rules

- Every field below maps to its snake_case wire key, unless explicitly stated otherwise. `MasterTimeout string` -> `master_timeout`, `Timeout string` -> `timeout`. Optional strings emit only when nonempty; no client defaults.
- New boolean fields are `*bool`, serialized by `setBoolParam`; lists are `[]string`, serialized by `setListParam`; `*int` uses intStr, `*int64` uses int64Str; pointers retain explicit zero. Fractional number: `*float64`, `strconv.FormatFloat(*p, 'g', -1, 64)`.
- Reuse Refresh, ExpandWildcards, SortOrder. New enum SnapshotSort has constants start_time, duration, name, repository, index_count, shard_count, failed_shard_count in api/param_types.go. Storage is a spec string (not an invented closed enum); SlmPolicyFilter is string despite comma-separated semantics because the pinned descriptor is string. Do not change existing field types.
- Extend structs/ToMap in their existing files. New extension carriers/interfaces/methods go in `api/spec_extensions_model.go` and `api/spec_extensions.go`; new validated SpecRequest structs/Validate methods go in `api/options.go`. Reuse doReq/doReqRaw/headReq/buildRequest, normalizeBody, pathSegment, escapedList, mergeParams, validator and existing error helpers. Close test servers, avoid races in captured requests.
- Adapters `As<Group>ServiceSpecExtensions` reject nil, typed-nil, unsupported custom services; default implementations have compile-time assertions. Register the extension, not the deficient old wrapper. Preserve old nil-param behavior; delegate old methods where equivalent. Do not delegate the text-structure legacy routes to new extensions.
- Common new shared carriers, each with Common *CommonParams and nil-safe ToMap: `MasterTimeoutSpecParams {MasterTimeout string}`, `TimeoutSpecParams {Timeout string}`, `MasterTimeoutAndTimeoutSpecParams {MasterTimeout, Timeout string}`. Only methods with those exact local param sets use them. Endpoint-specific carriers include other existing params; never discard them.
- All batches modify the common registration set **R**: `internal/speccheck/bindings.json`, generated `registry_generated.go`, generated `types_generated.go` (only if regeneration changes it), reviewed `baseline.go`, `README.md`, and `integrity_test.go` for targeted serialization/source regressions as needed. No manual generated-file edits; no policy/integrity weakening.
- Test file convention **T(group)**: `api/<group>_service_test.go` for table-driven transport; `api/<group>_model_test.go` for carrier tests (create if absent). Extensions also modify `api/spec_extensions_test.go`. Each batch below identifies source files in addition to R and these exact test paths.

## 3. Ordered independently green batches

Each numbered batch is a commit; split into its named carrier subgroups if necessary, keeping every subcommit independently compiled, audited, and tested. Shared carrier work precedes single-endpoint work. No commit leaves registry/baseline overlap or partial query coverage. Cumulative counts assume all preceding batches completed.

### B1 — shared timeout extension families (39; cumulative 197/315)

Files: R; api/spec_extensions.go, spec_extensions_model.go, options.go; api/ccr_service.go, slm_service.go, ilm_service.go, ilm_model.go, enrich_service.go, features_service.go, watcher_service.go, xpack_service.go; T(ccr/slm/ilm/enrich/features/watcher/xpack); spec_extensions_test.go.

- **ccr (13):** delete_auto_follow_pattern, follow, follow_info, follow_stats, forget_follower, get_auto_follow_pattern, pause_auto_follow_pattern, pause_follow, put_auto_follow_pattern, resume_auto_follow_pattern, resume_follow, stats, unfollow. MasterTimeoutSpecParams for all except follow/follow_stats/forget_follower/stats. TimeoutSpecParams for follow_stats/forget_follower; both carrier for stats. Direct `CcrFollowParams.MasterTimeout string` for follow. Body+selector methods PutAutoFollowPattern, ResumeFollow, ForgetFollower use WithRequest with current name/index/body plus Params; other carrier-less methods use WithParams. Preserve paths and required name/index checks; named/all list routes tested.
- **slm (9):** delete_lifecycle, execute_lifecycle, execute_retention, get_lifecycle, get_stats, get_status, put_lifecycle, start, stop. All share MasterTimeoutAndTimeoutSpecParams; PutLifecycleWithRequest, others WithParams. Bind policy_id from existing singular/list selector.
- **ilm (7):** delete_lifecycle, explain_lifecycle, get_lifecycle, migrate_to_data_tiers, put_lifecycle, start, stop. Both timeout carrier for delete/get/put/start/stop. PutLifecycleWithRequest; remaining simple methods WithParams. GetLifecycle policy part already reaches policies argument 1; test named/all routes, no invented selector. Direct IlmExplainParams.MasterTimeout. MigrateToDataTiersWithParams uses MasterTimeoutSpecParams and existing dryRun *bool as QueryArgs[dry_run] argument 1 (test true/false/nil).
- **enrich (5):** delete_policy, execute_policy, get_policy, put_policy, stats. Direct EnrichWaitForCompletionParams.MasterTimeout for execute; MasterTimeoutSpecParams on other extensions, PutPolicyWithRequest, others WithParams.
- **features (2):** get_features/reset_features share MasterTimeoutSpecParams via WithParams; preserve GET /_features and POST /_features/_reset.
- **watcher.start/stop (2)** and **xpack.usage (1):** MasterTimeoutSpecParams via WithParams.

### B2 — Security shared refresh and accurate API-key fields (15; cumulative 212/300)

Files: R; api/security_service.go, spec_extensions.go, spec_extensions_model.go, options.go; T(security); spec_extensions_test.go.

- Shared `SecurityRefreshSpecParams {Common *CommonParams; Refresh Refresh}` for **12** endpoints: change_password, create_api_key, delete_privileges, delete_role, delete_role_mapping, delete_user, disable_user, enable_user, put_privileges, put_role, put_role_mapping, put_user. Wire refresh via string conversion, including true/false/wait_for.
- Body+selector and application+name+params methods use WithRequest; CreateAPIKey/PutPrivileges (body+params) use WithParams; singular selector methods use WithParams. Preserve authenticated ChangePassword empty-username route. Request structs retain old argument semantics and required checks.
- **get_api_key:** actual GetAPIKey, existing SecurityGetAPIKeyParams; add WithLimitedBy *bool -> with_limited_by. `queryFields: {"id":"ID"}`; do not add duplicate Id. Preserve legacy WithLimitedPrivileges -> with_limited_privileges and WithProfileUid.
- **get_user:** GetUserWithParams and `SecurityGetUserSpecParams {Common; WithProfileUid *bool}`; username list named/all routes.
- **has_privileges:** HasPrivilegesWithRequest, `SecurityHasPrivilegesSpecRequest {User string; Body any}`; optional User selects POST /_security/user/{user}/_has_privileges, empty selects authenticated route. Required body validation on extension only. No params carrier needed.

### B3 — existing shared-carrier wins (15; cumulative 227/285)

Files: R; api/indices_model.go, dangling_indices_service.go, license_service.go; T(indices/dangling_indices/license). Transport changes only if required to send existing carrier; no extensions needed here.

| Endpoints | Canonical carrier additions | Binding detail |
| --- | --- | --- |
| indices.clone/shrink/split (3) | ShrinkParams.MasterTimeout string | index = request.Source, target = request.Target; keep aliases/signatures |
| indices.open/close (2) | IndicesOpenParams.MasterTimeout string | index = indices argument 1; CloseParams alias |
| indices.put_alias/delete_alias (2) | AliasParams.MasterTimeout string | put: name=request.Alias, index=request.Index; delete: names/indices request fields |
| indices.get_alias/exists_alias (2) | GetAliasParams.MasterTimeout string | name=request.Names, index=request.Indices |
| indices.simulate_index_template/simulate_template (2) | SimulateIndexTemplateParams.MasterTimeout string, IncludeDefaults *bool | keep SimulateTemplateParams alias; name=request.Name |
| dangling_indices.delete_dangling_index/import_dangling_index (2) | DanglingIndicesImportParams.MasterTimeout, Timeout string | index_uuid=argument 1; accept_data_loss existing *bool/tag unchanged |
| license.post/post_start_basic (2) | LicensePostParams.MasterTimeout, Timeout string | preserve Acknowledge and different routes/response types |

Regression: a single MasterTimeout change must pass live serialization and independent transport cases on every alias/user; explicit false accept_data_loss continues to serialize without introducing new validation.

### B4 — remaining indices carriers and four missing-carrier methods (27; cumulative 254/258)

Files: R; api/indices_model.go, indices_service.go, param_types.go (only shared enum additions if not already landed), spec_extensions.go, spec_extensions_model.go, options.go; T(indices); spec_extensions_test.go.

Direct additions and exact endpoint-to-carrier matrix:

| Endpoint (indices.) | Carrier | Additions |
| --- | --- | --- |
| add_block | AddBlockParams | MasterTimeout string |
| clear_cache | IndicesClearCacheParams | Request *bool |
| create | IndicesCreateParams | MasterTimeout, Timeout string |
| create_data_stream | IndicesCreateDataStreamParams | MasterTimeout string |
| delete | IndicesDeleteParams | MasterTimeout, Timeout string |
| delete_data_stream | IndicesDeleteDataStreamParams | MasterTimeout string |
| delete_index_template | IndicesDeleteIndexTemplateParams | MasterTimeout string |
| delete_template | IndicesDeleteTemplateParams | MasterTimeout string |
| get | IndicesGetParams | MasterTimeout string |
| get_data_stream | IndicesGetDataStreamParams | MasterTimeout string; Verbose *bool |
| get_index_template | IndicesGetIndexTemplateParams | MasterTimeout string; IncludeDefaults *bool |
| get_mapping | IndicesGetMappingParams | MasterTimeout string |
| get_settings | GetSettingsParams | MasterTimeout string |
| get_template | IndicesGetTemplateParams | MasterTimeout string |
| put_index_template | PutIndexTemplateParams | MasterTimeout string |
| put_mapping | PutMappingParams | MasterTimeout string |
| put_settings | IndicesPutSettingsParams | MasterTimeout string; Reopen *bool |
| put_template | PutTemplateParams | MasterTimeout, Cause string |
| reload_search_analyzers | IndicesReloadAnalyzersParams | Resource string |
| resolve_cluster | IndicesResolveClusterParams | IgnoreThrottled *bool; Timeout string |
| rollover | RolloverParams | MasterTimeout string; Lazy *bool |
| update_aliases | IndicesUpdateAliasesParams | MasterTimeout string |

- Missing-carrier extensions **exists_template/exists_index_template** reuse their now-complete IndicesGetTemplateParams/IndicesGetIndexTemplateParams respectively via WithParams (flat_settings/local/master_timeout; extra get include_defaults is omitted by default). headReq already accepts params; retain 200/404/error semantics. **migrate_to_data_stream** uses both timeout carrier; **promote_data_stream** master-only carrier, both WithParams.
- **forcemerge:** actual ForceMerge. Add WaitForCompletion *bool to IndicesForceMergeParams with wire wait_for_completion. Register `IndicesServiceSpecExtensions.ForceMergeWithRequest` with SpecRequest {Indices []string; Params *IndicesForceMergeParams} and `IndicesForceMergeSpecResponse` retaining the existing BroadcastResponse shard data plus Task string. Both synchronous and {task:...} asynchronous responses must be retained/tested. Original ForceMerge keeps result signature; do not silently claim its old result represents async mode. Document extension use for false. No raw-response workaround.
- clear_cache spec also declares index as a query parameter: use the existing indices positional argument as both QueryArgs[index] and PathSources[index] **only after adding/testing actual index query emission** in ClearCache; reflection does not prove it. Prefer narrow buildRequest/query changes, not a new duplicate Index field. Verify all endpoint-local params, not only old reason clauses.

### B5 — cluster timeout carriers and a genuine route correction (15; cumulative 269/243)

Files: R; api/cluster_model.go, cluster_service.go, spec_extensions.go, spec_extensions_model.go; T(cluster); spec_extensions_test.go.

- Add MasterTimeout string to ClusterAllocationExplainParams (allocation_explain), ClusterDeleteComponentTemplateParams (delete_component_template), ClusterDeleteVotingConfigExclusionsParams (delete_voting_config_exclusions), ClusterGetComponentTemplateParams (get_component_template), ClusterGetSettingsParams (get_settings), ClusterHealthParams (health), ClusterPendingTasksParams (pending_tasks), ClusterVotingConfigExclusionsParams (post_voting_config_exclusions), PutComponentTemplateParams (put_component_template), ClusterPutSettingsParams (put_settings), ClusterRerouteParams (reroute), ClusterStateParams (state): **12** endpoints.
- Also ClusterGetComponentTemplateParams.IncludeDefaults *bool; ClusterHealthParams.ExpandWildcards ExpandWildcards; ClusterStatsParams.IncludeRemotes *bool for **stats**.
- **exists_component_template:** ExistsComponentTemplateWithParams reuses ClusterGetComponentTemplateParams (local/master_timeout) and headReq; preserve omission of other get-only options by default and document applicability.
- **info:** bind target to existing features []string argument 1; current /_cluster/info route is wrong. Correct default Info directly to GET /_info/{escapedList(features)}; empty features selects /_info/_all (not a non-spec empty route). Preserve interface/response type. Test _all, each target and multi-target lists. This is an approved behavior-changing bug fix, not just a naming alias.

### B6 — snapshot params, paging, and positional repository aliases (12; cumulative 281/231)

Files: R; api/snapshot_model.go, snapshot_service.go, param_types.go; T(snapshot). Extensions only if validation/transport evidence requires a stop-condition fallback.

- Add MasterTimeout string to SnapshotCreateRepositoryParams, SnapshotGetRepositoryParams, SnapshotDeleteRepositoryParams, SnapshotVerifyRepositoryParams, SnapshotCleanupRepositoryParams, SnapshotCreateParams, SnapshotRestoreParams, SnapshotStatusParams: **create_repository/get_repository/delete_repository/verify_repository/cleanup_repository/create/restore/status**.
- **clone:** add optional Params *MasterTimeoutSpecParams to SnapshotCloneRequest and pass it to existing doReq. Its /_snapshot/{repository}/{snapshot}/_clone/{target_snapshot} route already matches the pin. Do not "correct" it based on the older commentary.
- **delete:** add Params *SnapshotDeleteSpecParams to SnapshotDeleteRequest; carrier {Common; MasterTimeout string; WaitForCompletion *bool}. Existing Snapshot string can represent a comma list; test list transport. No changed required validation.
- **get:** SnapshotGetParams adds MasterTimeout string, Order SortOrder, FromSortValue string, After string, Verbose *bool, IndexNames *bool, Sort SnapshotSort, Size *int, Offset *int, SlmPolicyFilter string. Keep existing fields. Current empty Snapshots path hits get_repository instead: correct to /_snapshot/{repository}/_all; named/list snapshots stay escaped. Test exact endpoint routing for empty/named/multiple snapshots and all paging keys/zero values.
- **repository_analyze:** SnapshotRepositoryAnalyzeParams adds EarlyReadNodeCount *int, ReadNodeCount *int, RarelyAbortWrites *bool, RareActionProbability *float64. No master_timeout addition (not in this spec). Test probability 0 and 0.02 and zero counts.
- Six "repository" gaps are naming-only: create/verify/analyze/cleanup -> argument 1 name string; get/delete repository -> argument 1 names []string. Bind and test; do not add fake new selectors. Bind snapshot/target_snapshot from real request fields.

### B7 — document/search/async/EQL fields, safe dedicated mtermvectors (15; cumulative 296/216)

Files: R; api/document_model.go, document_service.go, search_model.go, async_search_model.go, eql_service.go, spec_extensions.go, spec_extensions_model.go, options.go; T(document/search/async_search/eql); spec_extensions_test.go.

- **index/create:** shared IndexParams (CreateParams alias) adds Routing string, RequireDataStream *bool, IncludeSourceOnError *bool. Correct Create directly to PUT /{index}/_create/{id}, replacing non-spec /{index}/_doc/{id}/_create. Test encoding and create conflict semantics. Existing fields/types/required validation unchanged.
- **bulk:** BulkParams.RequireDataStream/IncludeSourceOnError *bool. **update:** UpdateParams.IncludeSourceOnError *bool. **delete:** DeleteParams.WaitForActiveShards string. **delete_by_query/update_by_query:** respective DeleteByQueryParams/UpdateByQueryParams.AllowNoIndices *bool.
- **mtermvectors:** DocumentServiceSpecExtensions.MultiTermVectorsWithRequest; MtermvectorsSpecRequest {Index string; Body any; Params *MtermvectorsSpecParams}; new carrier embeds TermVectorsParams **by value**, adds Ids []string -> ids. Its ToMap calls the embedded ToMap and adds ids. Existing MtermvectorsParams alias remains unchanged; old method may delegate by copying its old params into embedded carrier. Test nil params, nil body with ids, root/index paths, two ids, and unchanged termvectors query omission.
- **field_caps:** FieldCapsParams.Filters/Types []string, IncludeEmptyFields *bool. **open_point_in_time:** OpenPITParams.AllowPartialSearchResults *bool (do not change KeepAlive validation). **search_shards:** SearchShardsParams.MasterTimeout string. **search_template:** SearchTemplateParams.ExpandWildcards ExpandWildcards.
- **async_search.submit:** AsyncSearchSubmitParams.RestTotalHitsAsInt, SeqNoPrimaryTerm, Version, CcsMinimizeRoundtrips *bool; MaxConcurrentShardRequests *int; DocvalueFields []string. Preserve existing Search-like fields; do not replace/re-embed the entire SearchParams to chase sharing.
- **eql.search:** EqlSearchParams.AllowPartialSearchResults/AllowPartialSequenceResults *bool. **eql.get:** GetWithParams and EqlGetSpecParams {Common; WaitForCompletionTimeout, KeepAlive string}; existing ID argument 1. Do not reuse EqlSearchParams's unrelated options.

### B8 — ingest/script/license/health straightforward remaining timeouts (10; cumulative 306/206)

Files: R; api/ingest_model.go, script_model.go, script_service.go, license_service.go, health_report_service.go, spec_extensions.go; T(ingest/script/license/health_report); spec_extensions_test.go.

- **ingest.delete_pipeline/get_pipeline/put_pipeline:** corresponding IngestDeletePipelineParams/IngestGetPipelineParams/IngestPutPipelineParams.MasterTimeout string.
- **delete_script/get_script/put_script:** corresponding ScriptDeleteParams/ScriptGetParams/ScriptPutParams.MasterTimeout string. put_script adds optional Context string to ScriptPutRequest for the path /_scripts/{Id}/{Context}; retain Params.Context for independent query context. Select path suffix only when request.Context nonempty. Test different path/query context values, empty route, IDs and escaping. No extension required for additive optional selector.
- **license.delete:** DeleteWithParams, both timeout carrier. **license.get:** LicenseGetParams.AcceptEnterprise *bool. **license.post_start_trial:** LicenseStartTrialParams.MasterTimeout string. Preserve LicensePostParams shared changes from B3.
- **health_report:** HealthReportParams.Timeout string. Preserve Time string and feature positional routes; timeout is not a rename of Time.

### B9 — small namespace fields and missing-carrier extensions (13; cumulative 319/193)

Files: R; api/ml_service.go, searchable_snapshots_service.go, transform_model.go, transform_service.go, search_application_service.go, nodes_model.go, xpack_service.go, spec_extensions.go, spec_extensions_model.go, options.go; T(ml/searchable_snapshots/transform/search_application/nodes/xpack); spec_extensions_test.go.

- **ml.delete_job/reset_job:** respective MlDeleteJobParams/MlResetJobParams.DeleteUserAnnotations *bool. **ml.put_job/put_datafeed:** PutJobWithRequest/PutDatafeedWithRequest with old ID/body and shared `MlPutSpecParams {Common; IgnoreUnavailable, AllowNoIndices, IgnoreThrottled *bool; ExpandWildcards ExpandWildcards}`. Do not add body-dependent validation that breaks old callers; spec applicability is documented (put_job options apply with datafeed_config).
- **searchable_snapshots.mount:** SearchableSnapshotsMountParams.MasterTimeout, Storage string. **clear_cache:** ClearCacheWithParams and SearchableSnapshotsClearCacheSpecParams {Common; IgnoreUnavailable, AllowNoIndices *bool; ExpandWildcards ExpandWildcards}; named/all index paths.
- **transform.delete_transform:** TransformDeleteParams.DeleteDestIndex *bool. **start_transform:** TransformStartParams.From string (time expression, not int). **schedule_now_transform:** ScheduleNowWithParams using TimeoutSpecParams.
- **search_application.list:** SearchApplicationListParams.Q string. **search_application.search:** SearchWithRequest using new request {Name string; Body any; Params *SearchApplicationSearchSpecParams}; params {Common; TypedKeys *bool}. Existing SearchResult response retained/tested.
- **nodes.stats:** NodesStatsParams.IncludeUnloadedSegments *bool; test node_id/metric/index_metric paths, not just ToMap.
- **xpack.info:** XpackInfoParams.Human, AcceptEnterprise *bool. Human must be a directly settable endpoint field, not an ignore or only nested Common.Human; endpoint field overrides Common on merge.

### B10 — Cat controlled format and pure path naming gaps (12; cumulative 331/181)

Files: R; api/cat_model.go, inference_service.go, spec_extensions.go, options.go; T(cat/inference/synonyms/watcher); spec_extensions_test.go. Synonyms/watcher source files need no transport changes for naming aliases.

- **cat.nodes:** CatNodesParams.IncludeUnloadedSegments *bool. **cat.plugins:** CatPluginsParams.IncludeBootstrap *bool. **cat.tasks:** CatTasksParams.Nodes []string, Timeout string, WaitForCompletion *bool. Keep legacy NodeId -> node_id behavior unchanged; Nodes is the actual v9 wire nodes. Do not duplicate promoted h/help/s/v/local/master_timeout/time.
- **cat.thread_pool:** already has patterns []string argument 1 transported by escapedList; bind thread_pool_patterns to it. All four Cat carriers already emit format=json. Add their exact names to baseline.go's catFormatEndpoints, keeping policy.go logic intact: ignored-params grows 21 -> 25. Do not add a caller-settable Format field or apply exception to cat.help. Controlled-value tests and real transport must prove JSON, including malicious/list patterns and named/default paths.
- **synonyms.delete_synonym_rule/get_synonym_rule/put_synonym_rule:** set_id=argument 1 id; rule_id=argument 2 ruleId. Existing routes already correct. Required body for put and IDs retain old semantics; transport tests cover GET/PUT/DELETE with both distinct IDs.
- **watcher.ack_watch/activate_watch/deactivate_watch:** watch_id=argument 1 id; ack's action_id=argument 2 actionId; test named/no-action ack variants. No fake WatchId field or extension needed.
- **inference.update:** add InferenceServiceSpecExtensions.UpdateWithRequest, InferenceUpdateSpecRequest {InferenceId string; TaskType string; Body any}. Use **PUT**, not current POST, on /_inference/{id}/_update or /_inference/{task_type}/{id}/_update. Correct old Update's verb directly to PUT or delegate with empty TaskType (approved behavior fix). No extension Params carrier: no query params in pin. Preserve response signature; test response decode/error behavior and both paths. Do not register unrelated inference methods.

### B11 — stable text-structure route extensions (2; final 333/179)

Files: R; api/spec_extensions.go, options.go; api/spec_extensions_model.go only if request body type is placed there; T(text_structure); spec_extensions_test.go. Existing api/text_structure_service.go methods remain unchanged.

- `TextStructureServiceSpecExtensions.FindFieldStructureWithRequest(ctx, *TextStructureFindFieldSpecRequest) (json.RawMessage, error)`: request {Index string; Field string; Params *TextStructureFindParams}; validate required Index/Field; **GET /_text_structure/find_field_structure**; emit index and field query values from request (QueryArgs at argument 1 fields Index/Field), existing carrier handles all other params including DocumentsToSample. No PathSources (spec has no parts). Use buildRequest/explicit query override so required selectors cannot be overwritten.
- `FindMessageStructureWithRequest(ctx, *TextStructureFindMessageSpecRequest) (json.RawMessage, error)`: request {Messages []string; Params *TextStructureFindParams}; require nonempty messages on the new API; **POST /_text_structure/find_message_structure** with JSON body exactly {"messages": [...]} and no index path/query/body. The baseline's "index in the body" reason is incorrect; the pinned spec and body transport tests are authoritative. Keep original legacy method unchanged, and describe its route limitation in docs rather than silently translating index to messages.
- Reuse existing TextStructureFindParams/ToMap unchanged; existing Format/EcsCompatibility string types stay unchanged. Test nil params, every spec query field, messages with quoting/newlines/Unicode, required-input failures before network, Content-Type, raw response and HTTP/context errors; prove the old methods still use legacy routes. Both endpoints stable/straightforward, approved for registration.
- If either cannot pass honest route/body evidence, retain its exact original baseline entry pending review (document factual discrepancy separately), omit its binding, and revise final counts by one each. Never reinterpret index as messages to reach the count.

## 4. Per-endpoint registration loop — mandatory, including intermediate failure

For every manifest row (including siblings unlocked by a shared carrier):

1. Read inventory and pinned JSON **all** local params, union of path parts, routes/methods/body; inspect concrete method/helper. Record actual evidence and transport test names. Reasons are seeds, not exhaustive truth (clear_cache.index, Create route, Cluster.Info route, Inference.Update verb are counterexamples).
2. Add fields/ToMap, or the approved extension/request/route repair. Compile and add table-driven transport/carrier tests first. Never "fix" field list/source metadata to reach data not actually transported.
3. Add one reviewed bindings.json entry: actual service/method, original qualified carrier, exact params source, positional QueryArgs, exact PathSources, genuine initialism QueryFields overrides, and concrete route/test evidence. Input indices include ctx as 0, do not include a receiver; prefer explicit argument 1 + field Params for requests.
4. Run `go generate ./internal/speccheck`, then `go run ./internal/speccheck -spec "$SPEC_DIR"`. With the baseline still present the audit **must exit 1 for overlap**; inspect and fix any param/path/integrity failures. Implemented counter increments before validation and alone is not proof of coverage.
5. Run integrity/generator tests and the endpoint's transport tests. Positional QueryArgs bypass live ToMap checking; transport tests are mandatory evidence for each one. Path reflection likewise proves reachability, not routing.
6. Only after all real checks pass, remove that exact knownMissingEndpoints record and its README backlog row; update counts/edit matrix. Run audit again, now requiring exit 0. Stale baseline, metadata drift, absent/excluded endpoint, overlap, removed serialization keys remain failures.
7. Repeat for siblings. End the batch with full package/all-source checks; commit API changes/tests/binding/generated outputs/baseline/docs together. If it fails, keep the row baselined, remove the proposed binding, preserve an evidence-specific reason where appropriate, and do not claim completion.

Do not change generator input rules, feature-flag allowlist, ignores other than the four exact Cat format additions, spec loader behavior, or runtime integrity semantics. A newly registered endpoint must never survive in baseline; generation never writes baselines. No endpoint-local param/path debt suppression.

## 5. Tests and validation gates

- Every new field/carrier: nil receiver, zero/nil omission, explicit false/true, integer zero/nonzero, two-element list, string/enum wire values; merge precedence with Common; no accidental defaults. Shared-carrier regression tables enumerate **all** endpoints consuming a changed alias/carrier, plus unchanged non-target consumers.
- Every new registration has table-driven real HTTP evidence for every path/positional source, with distinct sentinels to detect swaps. Cover optional/default/named/list variants, escaped segment/list values (`/`, `?`, `#`, spaces, Unicode), method, query keys/values, body/content type, result decoding, relevant 404 semantics and errors. A static route-string assertion or ToMap-only test is not enough.
- Every new extension: default/nil/typed-nil/custom adapter cases; compile original interface consumers; nil/default params, required-input errors with no network, context cancellation, HTTP/network/malformed-response errors where decoded, old wrapper behavior. Do not add automatic retries for new methods here.
- Existing `TestValidateRegistryProductionEntries`, `TestGeneratedFieldsMatchReflection`, `TestCommittedGeneratedMatches`, policy overlap/staleness tests remain gates. Add targeted live-serialization tests for shared MasterTimeout carriers, Index/Create aliases, Cat promoted fields, mtermvectors embedding, and ID override. Keep removal-of-wire-key/field/source mutation tests failing closed. Do not lower coverage/lint thresholds.
- No local upstream JSON files committed. Fixtures are minimal authored test JSON, and mutation tests use temporary copies only, never alter `/tmp/es-spec` or the pin.

Implementation agent commands (this planner did not execute them):

```sh
cd /projects/elasticsearch
git status --porcelain=v1 --untracked-files=all
git branch --show-current
git rev-parse HEAD
git -C /tmp/es-spec rev-parse HEAD
SPEC_DIR=/tmp/es-spec/rest-api-spec/src/main/resources/rest-api-spec/api
go run ./internal/speccheck -spec "$SPEC_DIR" -inventory > /tmp/speccheck-before.json
go run ./internal/speccheck -spec "$SPEC_DIR"

# Each registration/batch, following the intermediate-overlap workflow above:
go generate ./internal/speccheck
go test -count=1 ./api ./internal/speccheck/...
go run ./internal/speccheck -spec "$SPEC_DIR"
go build ./...
go vet ./...
go test -race -count=1 ./...
dagger call --src . lint
dagger call --src . spec-check

# Final complete gates; Test/TestAcceptance return files, force evaluation/export:
dagger call --src . test export --path /tmp/speccheck-unit-coverage.out
dagger call --src . test-acceptance export --path /tmp/speccheck-acceptance-coverage.out
dagger call --src . ci export --path /tmp/speccheck-ci-output
```

Use the existing project Dagger CLI version/commands; verify file-returning exports with the local schema rather than treating an unevaluated object as a passed stage. Full `ci --ci` with upload/push is run in authorized CI with its existing secrets, not locally without credentials. Observe PR #1 required checks actually green.

Byte stability: after a green batch, copy both generated files to /tmp, run generation again, compare with `cmp`; `go test -count=1 ./internal/speccheck/generator` also compares generation to tracked outputs. Inspect git diff --check/stat and scoped changes, run gofmt on touched Go files before final generation. Confirm baseline membership and ignored counts separately from command success.

## 6. Commit, rollout, risks, and stop conditions

- Commit B1–B11 on the existing branch, messages `feat(api): complete <batch> spec bindings`; split a batch by shared carrier/service if tests/ownership need smaller units. R files and extension interfaces are integration-owned to avoid parallel conflicting edits. No new branch/rebase or CI edits planned.
- Re-run local audit and all generated/integrity tests after each commit. Run acceptance before/after direct Create/Cluster.Info/Inference.Update/Snapshot.Get route changes, not only at the end; add focused live acceptance cases for straightforward corrected routes when fixture capabilities permit. No requirement to provision ML/CCR/Security product scenarios for every endpoint; mock transport is the binding evidence, existing acceptance is the regression gate.
- Document new opt-in usage, unsupported custom-service errors, route corrections, and mtermvectors/ForceMerge extension behavior in internal/speccheck/README.md; update root README truthful counts if it states them. Update CONTRIBUTING only for genuinely new conventions (the existing extension/transport rules already apply); unrelated version wording is out of scope.
- Source compatibility caveat: additive exported fields can break downstream **unkeyed** struct literals. State keyed-literal expectation in release notes; do not claim universal Go source compatibility. Existing interfaces and keyed literals retain compatibility. If this is a release blocker, use a new SpecRequest/SpecParams extension instead of changing that carrier and keep the target row open until tests prove it.
- Embedding/initialisms: named Common is not promoted; aliases and unexported anonymous Cat containers are. Never duplicate fields to satisfy snakeToCamel. Use ID override, preserve actual APIKey/ForceMerge names, and compare AST to reflect.VisibleFields. For mtermvectors embed an exported value type, not a nil anonymous pointer.
- Shared changes: audit all consumers, nil/zero defaults and wire applicability. No global timeout/refresh fields on CommonParams. Dedicated mtermvectors extension prevents ids leaking to termvectors; security preserves legacy wire fields; template get carriers reused by exists are documented endpoint-specific capabilities, not forced query defaults.
- Validation: never add required timeout/refresh/optional selectors to old structs or change old body validation. Add new required validation only on extensions. If existing Validate rejects formerly valid calls after an additive edit, stop and move to SpecRequest/SpecParams or retain baseline with concrete evidence; do not loosen unrelated validation.
- Ambiguous transport/response: all routes/verbs/required parts must be tested against pin. Correct proven defects, not heuristic aliases. Preserve response types except the approved new ForceMerge extension response. If optional async params would discard task responses elsewhere, add a dedicated extension retaining that response or keep the endpoint baselined with evidence; no fabricated coverage through successful unmarshalling of `{}`.
- Runtime gate is intentionally not a route/body proof: live ToMap and reflection stay authoritative for their domains; transport tests supply the missing proof. No extension introduced solely to evade runtime serialization checks.
- Subagent/time budget: delegate bounded carrier/service subgroups with frozen manifest, concrete files and required tests; check in partial **green complete endpoint batches**, never bulk speculative bindings. Return unresolved rows and evidence when an agent stops; integration owner alone edits shared R outputs and counts. A timeout is not permission to delete baseline debt or declare a batch done.
- Rollback: revert a batch's API/bindings/generated outputs/tests/baseline removals/docs together; restore only its previously reviewed debt entries. Never roll back by disabling ratchet/integrity, changing pin, expanding ignores, or relabeling unverified implementation as supported.

## Definition of done

1. Frozen target set fully disposed: 175 newly verified registrations, no target baseline overlap, no generic binding-incomplete/route-mismatch reason remains. If a documented stop condition keeps a target baselined, the effort is **partial**, with truthful revised counts `implemented=333-D`, `known-missing=179+D` and an explicit list of D deferred names/reasons; never silently call that full completion.
2. Residual 179 entries are the untouched 45 absent-service/134 absent-method product backlog (experimental metadata retained); no methods from this residual set are implemented or opportunistically aliased.
3. Audit exits 0 on clean pinned v9.0.0 with checked=512, implemented=333, known-missing=179, ignored-params=25 and no param/path/integrity/policy drift. Registry conditional entries are retained, not included in these checked totals.
4. Generation is byte-stable and does not rewrite debt/ignores. Every carrier/source is real and caller-settable, every positional/path claim has transport evidence, shared changes have consumer regression coverage.
5. `go test -race ./...`, Dagger lint/spec-check/test/test-acceptance and the full existing CI stages pass; PR #1 CI is observed green. No committed spec clone, dependency/CI-policy changes, unrelated edits, or source implementation by this planning agent.

No unresolved product-policy questions remain. Field/route cross-checks in the frozen manifest are execution gates, not authorization to weaken coverage or expand product scope.
