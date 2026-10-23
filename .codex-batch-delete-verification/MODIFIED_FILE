<template>
  <el-config-provider :locale="zhCn">
    <section class="page-stack pool-page">
      <SectionHeader
        title="邮箱池"
        description="集中管理库存、项目接入和邮箱领取。"
      >
        <template #actions
          ><el-button :loading="loading" @click="refresh(true)"
            >刷新</el-button
          ></template
        >
      </SectionHeader>
      <RequestAlert
        v-if="error"
        :error="error"
        closable
        @close="error = null"
      />
      <el-tabs v-model="tab" @tab-change="refreshTab">
        <el-tab-pane label="邮箱库存" name="members">
          <div class="pool-toolbar">
            <el-select
              v-model="memberPurpose"
              placeholder="用途：Adobe"
              filterable
              allow-create
              @change="
                memberPage = 1;
                loadMembers();
              "
            >
              <el-option v-for="p in purposes" :key="p" :label="purposeName(p)" :value="p" />
            </el-select>
            <el-select
              v-model="memberAccountId"
              placeholder="全部主号"
              clearable
              filterable
              @change="
                memberPage = 1;
                loadMembers();
              "
            >
              <el-option
                v-for="account in mainAccounts"
                :key="account.id"
                :label="account.email"
                :value="account.id"
              />
            </el-select>
            <el-select
              v-model="memberState"
              placeholder="全部状态"
              clearable
              @change="
                memberPage = 1;
                loadMembers();
              "
            >
              <el-option
                v-for="s in ['available', 'leased', 'used', 'paused']"
                :key="s"
                :label="stateName(s)"
                :value="s"
              />
            </el-select>
            <el-input
              v-model="search"
              placeholder="搜索邮箱"
              clearable
              @keyup.enter="
                memberPage = 1;
                loadMembers();
              "
            />
            <el-button
              @click="
                memberPage = 1;
                loadMembers();
              "
              >搜索</el-button
            >
            <el-button type="primary" @click="openEnroll"
              >加入已有邮箱</el-button
            >
            <el-button
              type="danger"
              plain
              :loading="poolDeleteLoading"
              :disabled="busy || poolDeleteConfirming || poolDeletionState.recovering || poolDeletionState.uncertain || poolDeletionState.cancelling"
              @click="openPoolDeleteAction"
              >{{ poolDeletionActive ? "查看删除进度" : "批量删除当前筛选" }}</el-button
            >
          </div>
          <div class="pool-table-scroll" role="region" aria-label="邮箱库存列表" tabindex="0"><el-table
            :data="members.items"
            v-loading="loading"
            class="pool-table"
            empty-text="暂无库存。新创建且确认成功的邮箱会自动入池；已有邮箱可手动加入。"
          >
            <el-table-column label="序号" width="64" align="center"><template #default="{ $index }">{{ (memberPage - 1) * 50 + $index + 1 }}</template></el-table-column>
            <el-table-column label="邮箱" min-width="245">
              <template #default="{ row }">
                <div>{{ row.address }}</div>
                <small v-if="row.account_email" class="pool-member-account">{{ row.account_email }}</small>
              </template>
            </el-table-column>
            <el-table-column label="用途" width="100"><template #default="{ row }">{{ purposeName(row.purpose) }}</template></el-table-column>
            <el-table-column prop="group_name" label="分组" min-width="110" />
            <el-table-column label="状态" width="135"
              ><template #default="{ row }"
                ><el-tag :type="stateType(row.state)">{{
                  stateName(row.state)
                }}</el-tag
                ><small v-if="!row.enabled"> 主号/邮箱停用</small></template
              ></el-table-column
            >
            <el-table-column label="更新时间" min-width="175"
              ><template #default="{ row }">{{
                date(row.updated_at)
              }}</template></el-table-column
            >
            <el-table-column label="操作" min-width="185"
              ><template #default="{ row }">
                <el-button
                  link
                  type="primary"
                  @click="showCredentials(row.alias_id)"
                  >凭据</el-button
                >
                <el-button
                  v-if="row.state === 'available' || row.state === 'paused'"
                  link
                  :disabled="busy"
                  @click="setMember(row)"
                  >{{
                    row.state === "paused" ? "恢复分配" : "暂停分配"
                  }}</el-button
                >
              </template></el-table-column
            >
          ></el-table></div>
          <el-pagination
            v-model:current-page="memberPage"
            :page-size="50"
            :total="members.total"
            layout="total, prev, pager, next"
            @current-change="loadMembers"
          />
        </el-tab-pane>

        <el-tab-pane label="领取记录" name="leases">
          <div class="pool-toolbar">
            <el-select
              v-model="leasePurpose"
              placeholder="用途：Adobe"
              filterable
              allow-create
              @change="
                leasePage = 1;
                loadLeases();
              "
            >
              <el-option v-for="p in purposes" :key="p" :label="purposeName(p)" :value="p" />
            </el-select>
            <el-select
              v-model="leaseState"
              placeholder="全部状态"
              clearable
              @change="
                leasePage = 1;
                loadLeases();
              "
              ><el-option
                v-for="s in ['leased', 'used', 'released', 'expired']"
                :key="s"
                :label="stateName(s)"
                :value="s"
            /></el-select>
            <el-select
              v-model="clientFilter"
              placeholder="全部项目"
              clearable
              @change="
                leasePage = 1;
                loadLeases();
              "
              ><el-option
                v-for="c in clients"
                :key="c.id"
                :label="c.project"
                :value="c.id"
            /></el-select>
          </div>
          <div class="pool-table-scroll" role="region" aria-label="领取记录列表" tabindex="0"><el-table
            :data="leases.items"
            v-loading="loading"
            class="pool-table"
            empty-text="项目通过 API 领取邮箱后，记录会显示在这里。"
          >
            <el-table-column label="序号" width="64" align="center"><template #default="{ $index }">{{ (leasePage - 1) * 50 + $index + 1 }}</template></el-table-column>
            <el-table-column prop="address" label="邮箱" min-width="230" />
            <el-table-column prop="project" label="项目" min-width="110" />
            <el-table-column label="用途" width="100"><template #default="{ row }">{{ purposeName(row.purpose) }}</template></el-table-column>
            <el-table-column label="状态" width="100"
              ><template #default="{ row }"
                ><el-tag :type="stateType(row.state)">{{
                  stateName(row.state)
                }}</el-tag></template
              ></el-table-column
            >
            <el-table-column label="领取 / 到期时间" min-width="185"
              ><template #default="{ row }"
                ><div>{{ date(row.created_at) }}</div>
                <small>{{
                  row.state === "used"
                    ? (row.immediate_use ? "领取即已使用，保留分配" : "已确认使用，保留分配")
                    : date(row.expires_at)
                }}</small></template
              ></el-table-column
            >
            <el-table-column
              prop="request_id"
              label="请求编号"
              min-width="145"
              show-overflow-tooltip
            />
            <el-table-column label="操作" min-width="175"
              ><template #default="{ row }">
                <el-button
                  v-if="row.state === 'leased'"
                  link
                  type="primary"
                  :disabled="busy"
                  @click="actLease(row, 'commit')"
                  >确认使用</el-button
                >
                <el-button
                  v-if="row.state === 'leased' || (row.state === 'used' && row.immediate_use)"
                  link
                  type="danger"
                  :disabled="busy"
                  @click="actLease(row, 'release')"
                  >释放</el-button
                >
                <el-button
                  v-if="row.state === 'used'"
                  link
                  @click="showCredentials(row.alias_id)"
                  >凭据</el-button
                >
              </template></el-table-column
            >
          ></el-table></div>
          <el-pagination
            v-model:current-page="leasePage"
            :page-size="50"
            :total="leases.total"
            layout="total, prev, pager, next"
            @current-change="loadLeases"
          />
        </el-tab-pane>

        <el-tab-pane label="项目 API Key" name="clients">
          <div class="pool-toolbar">
            <el-input
              v-model="projectName"
              placeholder="项目名称，如 my-app"
              maxlength="80"
            /><el-button type="primary" :loading="busy" @click="createClient"
              >创建项目 Key</el-button
            >
          </div>
          <p class="pool-hint">
            Key 用于鉴权，只能读取该 Key 的领取记录；多个项目可共用一把 Key，通过 purpose 区分用途。项目 Key
            仅在创建或轮换时展示。
          </p>
          <div class="pool-table-scroll" role="region" aria-label="项目 API Key 列表" tabindex="0"><el-table
            :data="clients"
            v-loading="loading || clientsLoading"
            class="pool-table"
            empty-text="创建第一个项目，将 Key 配置到调用程序。"
          >
            <el-table-column prop="project" label="项目" min-width="150" />
            <el-table-column label="状态" width="100"
              ><template #default="{ row }"
                ><el-tag :type="row.enabled ? 'success' : 'info'">{{
                  row.enabled ? "启用" : "停用"
                }}</el-tag></template
              ></el-table-column
            >
            <el-table-column label="创建时间" min-width="185"
              ><template #default="{ row }">{{
                date(row.created_at)
              }}</template></el-table-column
            >
            <el-table-column label="操作" min-width="180"
              ><template #default="{ row }"
                ><el-button
                  link
                  type="primary"
                  :disabled="busy"
                  @click="updateClient(row, true)"
                  >轮换 Key</el-button
                ><el-button
                  link
                  :disabled="busy"
                  @click="updateClient(row, false)"
                  >{{ row.enabled ? "停用" : "启用" }}</el-button
                ></template
              ></el-table-column
            >
          ></el-table></div>
          <p class="pool-hint">
            停用项目会停止其邮箱池 API
            访问。已发放的单邮箱凭据可在「隐私邮箱」中单独轮换或停用。
          </p>
        </el-tab-pane>

        <el-tab-pane label="库存与定时创建" name="accounts">
          <div class="pool-toolbar">
            <span>领取即标记已使用</span>
            <el-switch
              :model-value="immediateUsed"
              :disabled="busy || loading"
              active-text="开启"
              inactive-text="关闭"
              @change="saveImmediateUsed"
            />
          </div>
          <p class="pool-hint">
            开启后，项目调用 claim 会立即得到 used 状态，不会因领取期限到期自动回池；确定未使用时调用 release 主动释放；仅在没有其他用途占用时轮换邮箱凭据。关闭后，按原有 leased → commit / release 流程。切换只影响新领取，已有记录保留原模式。
          </p>
          <p class="pool-hint">下方空闲库存和定时创建目标按“没有任何用途占用”的邮箱统计。上方库存状态按所选用途计算：Adobe 已用不影响 GPT 首次领取。暂停/恢复分配仍是全局操作。
          </p>
          <p class="pool-hint">
            本系统新创建并确认成功、凭据完整的匿名邮箱会直接进入库存，无需开启开关。下方开关仅用于将以后目录同步发现的新地址自动入池；已有邮箱在「邮箱库存」中手动选择加入。空闲目标为
            0
            时按计划创建；达到正数目标时跳过创建，领取后自动恢复。手动批量创建和单次探测均不使用本地创建额度，成功后不额外等待；仍受 Apple 实际响应影响。自动计划每主号滚动 1 小时最多 23 次：Apple Account 新通道 19 次、iCloud Web 旧通道 4 次，至少间隔 2 分钟并错峰执行。仅在明确限流且未产生地址时切换通道；结果不确定、含 HME 或目录确认中的尝试不会自动补造。明确限流默认暂停对应通道 1 小时，更长 Retry-After 以 Apple 为准。本地上限不代表 Apple 官方配额保证。
          </p>
          <el-table
            :data="accounts"
            empty-text="先到「主号管理」添加账号，并完成 Apple 登录。"
          >
            <el-table-column label="主号" min-width="240"
              ><template #default="{ row }"
                ><router-link
                  :to="{
                    name: 'account-detail',
                    params: { id: row.account_id },
                  }"
                  >{{ row.email }}</router-link
                ></template
              ></el-table-column
            >
            <el-table-column label="同步新地址入池" width="150"
              ><template #default="{ row }"
                ><el-switch v-model="row.auto_enroll" /></template
            ></el-table-column>
            <el-table-column label="空闲目标" width="165"
              ><template #default="{ row }"
                ><el-input-number
                  v-model="row.target"
                  :min="0"
                  :max="10000"
                  :step="10"
                  controls-position="right"
                  style="width: 140px" /></template
            ></el-table-column>
            <el-table-column prop="available" label="空闲" width="75" />
            <el-table-column label="定时创建" min-width="200"
              ><template #default="{ row }"
                ><el-switch
                  :model-value="row.auto_create"
                  :disabled="busy"
                  @change="toggleAutoCreate(row, $event)"
                />
                <div>
                  <small>{{
                    row.auto_create
                      ? "下次检查：" + date(row.next_run_at)
                      : "已关闭"
                  }}</small>
                </div>
                <small v-if="row.creation_status === 'cooldown'" class="pool-cooldown">创建暂停中；请在主号详情查看是本地预算等待还是 Apple 限流</small>
                <small v-else-if="row.last_error" class="pool-error">{{
                  row.last_error
                }}</small></template
              ></el-table-column
            >
            <el-table-column label="操作" width="90"
              ><template #default="{ row }"
                ><el-button
                  type="primary"
                  link
                  :disabled="busy"
                  @click="saveAccount(row)"
                  >保存</el-button
                ></template
              ></el-table-column
            >
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="API 接入" name="api">
          <div class="pool-api">
            <h3>1. 领取邮箱及完整凭据</h3>
            <p>
              使用项目 Key。request_id
              请在请求前保存，网络重试使用相同编号和参数。每次领取 1–50
              个邮箱，库存不足时整批保持原状。
            </p>
            <pre>{{ claimExample }}</pre>
            <h3>2. 查询本次领取后的新验证码</h3>
            <pre>{{ codeExample }}</pre>
            <p>
              data.success 为 true 时读取 data.otp；no_code 时等待至少 3 秒再请求。
              同一邮箱共用取件限流，429 时按 Retry-After 重试。可以追加
              after=RFC3339 时间，只查询更晚的邮件。
            </p>
            <h3>3. 释放或确认使用</h3>
            <pre>
POST /api/v1/pool/leases/LEASE_ID/commit
POST /api/v1/pool/leases/LEASE_ID/release
POST /api/v1/pool/leases/LEASE_ID/renew
续期 JSON: {"ttl_seconds":1800}</pre
            >
            <p v-if="immediateUsed">当前已开启领取即已使用：claim 立即占用，不会到期回池；确定未使用时调用 release，释放时轮换邮箱整套凭据。commit 对已使用记录是幂等操作。所有请求均带项目 Bearer Key。</p>
            <p v-else>当前为租用模式：默认领取 30 分钟，支持 60–86400 秒；commit 保留分配，release 或到期回收会轮换邮箱整套凭据。所有请求均带项目 Bearer Key。</p>
            <h3>4. 单邮箱 API、IMAPS 和 OAuth</h3>
            <pre>
GET /api/v1/otp
Authorization: Bearer MAILBOX_API_KEY

POST /oauth2/v2.0/token
Content-Type: application/x-www-form-urlencoded
grant_type=refresh_token&amp;client_id=CLIENT_ID&amp;refresh_token=REFRESH_TOKEN</pre
            >
            <p>
              领取结果包含邮箱地址、API Key、IMAP 密码、client_id、refresh_token
              和取码路径。IMAPS
              使用邮箱地址作为用户名。已有邮件归档随邮箱保留；按本次任务取码请使用领取记录的
              code 接口。
            </p>
          </div>
        </el-tab-pane>
      </el-tabs>

      <el-dialog
        v-model="poolDeleteConfirmOpen"
        title="确认批量删除邮箱池结果"
        width="min(760px,94vw)"
        destroy-on-close
        :close-on-click-modal="false"
        :close-on-press-escape="!poolDeleteConfirming"
        :show-close="!poolDeleteConfirming"
      >
        <p class="pool-delete-warning">
          将按当前筛选结果从 Apple 永久删除这些隐私邮箱。同一主号内串行，不同主号最多两个执行名额，并按 Apple 限流策略等待；已发出的邮箱不会因取消而恢复。
        </p>
        <div class="pool-delete-filters">
          <strong>筛选条件</strong>
          <span v-for="filter in poolDeletePreview.filters" :key="filter.label">{{ filter.label }}：{{ filter.value }}</span>
        </div>
        <p class="pool-delete-count">匹配 {{ poolDeletePreview.total }} 个邮箱</p>
        <div class="pool-delete-preview" role="region" aria-label="待删除邮箱预览">
          <div v-for="row in poolDeletePreview.items" :key="row.alias_id" class="pool-delete-preview__row">
            <span>{{ row.address }}</span>
            <small>{{ row.account_email }}</small>
          </div>
          <p v-if="poolDeletePreview.total > poolDeletePreview.items.length" class="pool-hint">
            仅展示前 {{ poolDeletePreview.items.length }} 个，确认后将处理全部 {{ poolDeletePreview.total }} 个。
          </p>
        </div>
        <template #footer>
          <el-button :disabled="poolDeleteConfirming" @click="poolDeleteConfirmOpen = false">取消</el-button>
          <el-button
            type="danger"
            :loading="poolDeleteConfirming"
            :disabled="!poolDeletePreview.ids.length"
            @click="submitPoolDelete"
          >确认删除 {{ poolDeletePreview.total }} 个</el-button>
        </template>
      </el-dialog>

      <el-dialog
        v-model="poolDeleteProgressOpen"
        title="邮箱池批量删除进度"
        width="min(760px,94vw)"
        :close-on-click-modal="false"
      >
        <template v-if="poolDeletionJob">
          <div class="pool-delete-progress__header">
            <el-tag :type="poolDeletionJobType">{{ poolDeletionStatusLabel }}</el-tag>
            <span v-if="poolDeletionJob.requestId">请求编号：{{ poolDeletionJob.requestId }}</span>
          </div>
          <p role="status" aria-live="polite">
            已处理 {{ poolDeletionJob.processed }} / {{ poolDeletionJob.requested }}；成功删除 {{ poolDeletionJob.deleted }}；失败/未执行 {{ poolDeletionJob.failed }}
            <span v-if="poolDeletionJob.deferred">（未执行 {{ poolDeletionJob.deferred }}）</span>
          </p>
          <el-progress :percentage="poolDeletionPercentage" :show-text="false" aria-label="邮箱池批量删除进度" />
          <p class="pool-hint">任务 {{ poolDeletionJob.jobId }} · 更新于 {{ date(poolDeletionJob.updatedAt) }}</p>
          <div v-if="poolDeletionWaits.length" class="pool-delete-waits" role="status" aria-live="polite">
            <strong>Apple 限流等待</strong>
            <span v-for="(wait, index) in poolDeletionWaits" :key="`${wait.accountId}:${wait.aliasId}:${index}`">
              主号 {{ wait.accountId }} · {{ ALIAS_DELETION_OPERATION_LABELS[wait.operation] }} · 下次 {{ formatTime(wait.retryAt, { seconds: true }) }}
            </span>
          </div>
          <div v-if="poolDeletionRecentFailures.length" class="pool-delete-failures">
            <strong>最近失败或待核对结果</strong>
            <span v-for="item in poolDeletionRecentFailures" :key="item.id">{{ item.address || `ID ${item.id}` }}：{{ formatAliasDeletionResultMessage(item) }}</span>
          </div>
        </template>
        <p v-else-if="poolDeletionState.recovering">正在恢复最近删除任务，请稍候。</p>
        <p v-else>当前没有可显示的删除任务。</p>
        <p v-if="isAliasDeletionJobActive(poolDeletionJob)" class="pool-hint">
          关闭此窗口不会停止后台任务；可以再次点击“查看删除进度”打开。
        </p>
        <template #footer>
          <el-button
            :loading="poolDeletionState.checking"
            :disabled="poolDeletionState.cancelling || poolDeletionState.submitting"
            @click="poolDeletionController.refresh()"
          >刷新进度</el-button>
          <el-button
            v-if="isAliasDeletionJobActive(poolDeletionJob)"
            type="warning"
            :loading="poolDeletionState.cancelling"
            :disabled="poolDeletionState.checking || poolDeletionState.submitting"
            @click="cancelPoolDelete"
          >取消任务</el-button>
          <el-button @click="poolDeleteProgressOpen = false">关闭</el-button>
        </template>
      </el-dialog>

      <el-dialog
        v-model="enrollOpen"
        title="加入已有邮箱"
        width="min(760px,94vw)"
        destroy-on-close
        :show-close="!enrollSubmitting"
        :close-on-click-modal="!enrollSubmitting"
        :close-on-press-escape="!enrollSubmitting"
      >
        <p>
          选择启用主号下已确认且凭据完整、当前可分配的邮箱加入池中。列表支持分页浏览，也可以搜索地址定位。
        </p>
        <RequestAlert v-if="error" :error="error" closable @close="error = null" />
        <p v-if="enrollProgress" class="pool-hint">{{ enrollProgress }}</p>
        <div class="pool-toolbar">
          <el-input
            v-model="enrollDraft"
            placeholder="搜索邮箱"
            :disabled="enrollLocked"
            @keyup.enter="applyEnrollSearch"
          /><el-button :disabled="enrollLocked" @click="applyEnrollSearch">搜索</el-button>
          <el-button :disabled="enrollLocked || !enrollItems.length" @click="selectCurrentPage">全选本页</el-button>
          <el-button :disabled="enrollLocked || !enrollItems.length" @click="invertEnrollPage">反选本页</el-button>
          <el-button :disabled="enrollLocked || !enrollTotal" :loading="enrollSelectingAll" @click="selectAllResults">全选搜索结果</el-button>
          <el-button :disabled="enrollLocked || !selected.length" @click="clearSelection">清空选择</el-button>
          <span class="pool-enrollment-count" aria-live="polite">已选 {{ selected.length }}</span>
        </div>
        <div ref="enrollSelectionContainer" @pointerdown.capture="enrollDrag.pointerDown">
          <el-table
            ref="enrollTable"
            row-key="id"
            :data="enrollItems"
            v-loading="enrollLoading || enrollSelectingAll"
            max-height="380"
            @selection-change="onEnrollSelectionChange"
        >
          <el-table-column
            type="selection"
            width="45"
            :selectable="(row) => !enrollLocked && isSelectable(row)"
          >
            <template #default="{ row }">
              <el-checkbox
                :model-value="selected.includes(row.id)"
                :disabled="enrollLocked || !isSelectable(row)"
                :data-selection-id="row.id"
                :data-selection-disabled="enrollLocked || !isSelectable(row)"
                :aria-label="`勾选 ${row.address}`"
                @change="setEnrollRow(row, $event)"
              />
            </template>
          </el-table-column>
          <el-table-column prop="address" label="邮箱" /><el-table-column
            prop="accountEmail"
            label="主号"
          />
          </el-table>
        </div>
        <div v-if="enrollTotal > 0" class="enroll-pagination">
          <span>共 {{ enrollTotal }} 条</span>
          <el-pagination
            background
            layout="prev, pager, next"
            size="small"
            :current-page="enrollPage"
            :page-size="enrollPageSize"
            :total="enrollTotal"
            :disabled="enrollLocked"
            :pager-count="5"
            @current-change="changeEnrollPage"
          />
        </div>
        <template #footer
          ><el-button :disabled="enrollSubmitting" @click="closeEnroll">取消</el-button
          ><el-button
            type="primary"
            :disabled="selected.length === 0 || enrollLocked"
            :loading="enrollSubmitting"
            @click="enroll"
            >加入 {{ selected.length }} 个邮箱</el-button
          ></template
        >
      </el-dialog>

      <el-dialog
        v-model="secretOpen"
        :title="secretTitle"
        width="min(730px,94vw)"
        @closed="secret = ''"
      >
        <el-input :model-value="secret" type="textarea" :rows="12" readonly />
        <template #footer
          ><el-button type="primary" @click="copySecret">复制</el-button
          ><el-button @click="secretOpen = false">关闭</el-button></template
        >
      </el-dialog>
    </section>
  </el-config-provider>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import zhCn from "element-plus/es/locale/lang/zh-cn";
import { apiRequest } from "../api/client.js";
import {
  cancelAliasDeletionJob,
  getAliasDeletionJob,
  getAliasPage,
  getAllAliases,
  getAllAccounts,
  getLatestAliasDeletionJob,
  startAliasDeletionJob,
} from "../api/admin.js";
import { mergePoolPageSelection, poolAliasSelectable, submitPoolEnrollment } from "../utils/poolEnrollment.js";
import { useAuth } from "../stores/auth.js";
import SectionHeader from "../components/SectionHeader.vue";
import RequestAlert from "../components/RequestAlert.vue";
import { copyText } from "../utils/clipboard.js";
import { createCheckboxDragSelection } from "../utils/checkboxDragSelection.js";
import {
  ALIAS_DELETION_OPERATION_LABELS,
  createAliasDeletionController,
  createAliasDeletionStorage,
  formatAliasDeletionResultMessage,
  isAliasDeletionJobActive,
  isAliasDeletionJobTerminal,
} from "../utils/aliasDeletionJob.js";
import { ADMIN_BASE_PATH } from "../utils/runtimePath.js";
import { formatTime } from "../utils/format.js";

const auth = useAuth();
const tab = ref("members"),
  loading = ref(false),
  busy = ref(false),
  error = ref(null);
const members = ref({ items: [], total: 0 }),
  leases = ref({ items: [], total: 0 }),
  clients = ref([]),
  accounts = ref([]),
  mainAccounts = ref([]);
const immediateUsed = ref(true);
const clientsLoaded = ref(false);
const clientsLoading = ref(false);
let clientsInFlight;
const memberState = ref(""),
  memberPurpose = ref("adobe"),
  memberAccountId = ref(""),
  leaseState = ref(""),
  leasePurpose = ref("adobe"),
  search = ref(""),
  clientFilter = ref(""),
  projectName = ref("");
const memberPage = ref(1),
  leasePage = ref(1);
const poolDeleteLoading = ref(false),
  poolDeleteConfirmOpen = ref(false),
  poolDeleteConfirming = ref(false),
  poolDeleteProgressOpen = ref(false),
  poolDeletePreview = ref({ filters: [], total: 0, ids: [], items: [] });
const enrollOpen = ref(false),
  enrollDraft = ref(""),
  enrollSearch = ref(""),
  enrollPage = ref(1),
  enrollPageSize = 50,
  enrollTotal = ref(0),
  enrollItems = ref([]),
  selected = ref([]),
  enrollTable = ref(null),
  enrollProgress = ref("");
const enabledAccountIds = ref(new Set());
const enrollLoading = ref(false), enrollSelectingAll = ref(false), enrollSubmitting = ref(false);
const enrollLocked = computed(() => busy.value || enrollLoading.value || enrollSelectingAll.value || enrollSubmitting.value);
let enrollTicket = 0;
let restoringEnrollSelection = false;
let enrollAbortController;
const enrollSelectionContainer = ref(null);
let enrollRestoreQueued = false;
const enrollDrag = createCheckboxDragSelection({
  getContainer: () => enrollSelectionContainer.value,
  getSelected: () => selected.value,
  isDisabled: () => enrollLocked.value || !enrollOpen.value,
  onChange: (ids) => {
    const disabled = new Set(enrollItems.value.filter((row) => !isSelectable(row)).map((row) => row.id));
    selected.value = ids.filter((id) => !disabled.has(id));
    queueEnrollRestore();
  },
});
function queueEnrollRestore() {
  if (enrollRestoreQueued) return;
  enrollRestoreQueued = true;
  void nextTick(async () => {
    try { await restoreEnrollSelection(); } finally { enrollRestoreQueued = false; }
  });
}
function setEnrollRow(row, checked) {
  if (enrollLocked.value || !isSelectable(row)) return;
  const ids = new Set(selected.value);
  if (checked) ids.add(row.id); else ids.delete(row.id);
  selected.value = [...ids];
  queueEnrollRestore();
}
async function invertEnrollPage() {
  if (enrollLocked.value) return;
  const ids = new Set(selected.value);
  for (const row of enrollItems.value.filter(isSelectable)) {
    if (ids.has(row.id)) ids.delete(row.id); else ids.add(row.id);
  }
  selected.value = [...ids];
  await restoreEnrollSelection();
}
const secretOpen = ref(false),
  secretTitle = ref(""),
  secret = ref("");
const purposes = ["common", "adobe", "gpt"];
const purposeName = (p) => ({ common: "通用", adobe: "Adobe", gpt: "GPT" })[p] || p || "通用";
const stateName = (s) =>
  ({
    available: "空闲",
    leased: "已领取",
    used: "已使用",
    paused: "暂停分配",
    released: "已释放",
    expired: "已到期",
  })[s] || s;
const stateType = (s) =>
  ({
    available: "success",
    leased: "warning",
    used: "primary",
    paused: "info",
    released: "info",
    expired: "info",
  })[s] || "info";
const date = (value) =>
  value ? new Date(value).toLocaleString("zh-CN", { hour12: false }) : "-";
const query = (params) =>
  new URLSearchParams(
    Object.entries(params).filter(([, v]) => v !== "" && v != null),
  ).toString();
const request = (path, method = "GET", body, signal) =>
  apiRequest(path, { method, body, signal, csrfToken: auth.state.csrfToken });
const poolDeletionState = ref({
  job: null,
  operationId: "",
  recovering: true,
  submitting: false,
  cancelling: false,
  checking: false,
  uncertain: false,
  error: null,
  blocked: true,
});
let poolDeletionController;
function makePoolDeletionController() {
  const username = auth.state.username;
  return createAliasDeletionController({
    startJob: startAliasDeletionJob,
    getJob: getAliasDeletionJob,
    getLatestJob: getLatestAliasDeletionJob,
    cancelJob: cancelAliasDeletionJob,
    storage: createAliasDeletionStorage(ADMIN_BASE_PATH, username),
    onChange(next) {
      if (auth.state.username !== username) return;
      poolDeletionState.value = next;
      if (isAliasDeletionJobTerminal(next.job)) {
        void loadMembers();
      }
    },
  });
}
poolDeletionController = makePoolDeletionController();
const poolDeletionJob = computed(() => poolDeletionState.value.job);
const poolDeletionActive = computed(() =>
  poolDeletionState.value.submitting || isAliasDeletionJobActive(poolDeletionJob.value),
);
const poolDeletionPercentage = computed(() => {
  const job = poolDeletionJob.value;
  return job?.requested ? Math.min(100, Math.max(0, job.processed / job.requested * 100)) : 0;
});
const poolDeletionRecentFailures = computed(() =>
  (poolDeletionJob.value?.results || []).filter((item) => !item.deleted).slice(-5).reverse(),
);
const poolDeletionWaits = computed(() =>
  poolDeletionJob.value?.status === "running" ? poolDeletionJob.value.waits || [] : [],
);
const poolDeletionJobType = computed(() => {
  if (poolDeletionState.value.uncertain || poolDeletionWaits.value.length || poolDeletionJob.value?.failed || poolDeletionJob.value?.status === "interrupted") return "warning";
  return poolDeletionJob.value?.status === "completed" ? "success" : "info";
});
const poolDeletionStatusLabel = computed(() => {
  if (poolDeletionState.value.submitting) return "正在提交";
  if (poolDeletionState.value.cancelling) return "正在取消";
  if (poolDeletionState.value.uncertain) return "结果待确认";
  if (poolDeletionState.value.recovering) return "恢复任务中";
  if (poolDeletionWaits.value.length) return "执行中（Apple 限流等待）";
  return { queued: "排队中", running: "执行中", completed: "已完成", interrupted: "已中断" }[poolDeletionJob.value?.status] || "查询中";
});
watch(() => auth.state.username, (username, previous) => {
  if (!username || username === previous) return;
  poolDeletionController.stop();
  poolDeletionController = makePoolDeletionController();
  poolDeletionState.value = poolDeletionController.getState();
  void poolDeletionController.start();
}, { flush: "sync" });
let poolRefreshTicket = 0;
let poolRefreshController;
let memberReadTicket = 0;
let leaseReadTicket = 0;
async function attempt(fn) {
  try {
    return await fn();
  } catch (e) {
    if (e.name !== "AbortError") error.value = e;
    return undefined;
  }
}
async function loadMembers(signal) {
  const ticket = ++memberReadTicket;
  await attempt(async () => {
    const result = await request(
      "/pool/members?" +
        query({
          state: memberState.value,
          purpose: memberPurpose.value,
          account_id: memberAccountId.value,
          q: search.value,
          limit: 50,
          offset: (memberPage.value - 1) * 50,
        }), "GET", undefined, signal instanceof AbortSignal ? signal : undefined,
    );
    if (ticket === memberReadTicket) members.value = result;
  });
}
function poolDeleteFilterSummary() {
  const account = mainAccounts.value.find((item) => String(item.id) === String(memberAccountId.value));
  return [
    { label: "用途", value: purposeName(memberPurpose.value) },
    { label: "主号", value: account?.email || "全部主号" },
    { label: "状态", value: memberState.value ? stateName(memberState.value) : "全部状态" },
    { label: "搜索", value: search.value.trim() || "全部邮箱" },
  ];
}
async function getAllFilteredPoolMembers() {
  const items = [];
  const snapshot = {
    state: memberState.value,
    purpose: memberPurpose.value,
    account_id: memberAccountId.value,
    q: search.value.trim(),
  };
  let offset = 0;
  let total = 0;
  do {
    const result = await request("/pool/members?" + query({ ...snapshot, limit: 200, offset }));
    const pageItems = Array.isArray(result?.items) ? result.items : [];
    items.push(...pageItems);
    total = Number(result?.total) || items.length;
    offset += pageItems.length;
    if (!pageItems.length) break;
  } while (items.length < total);
  return { items, total: items.length };
}
async function openPoolDeleteAction() {
  if (poolDeletionActive.value) {
    poolDeleteProgressOpen.value = true;
    return;
  }
  if (poolDeleteLoading.value || poolDeleteConfirming.value || poolDeletionState.value.recovering || poolDeletionState.value.uncertain || poolDeletionState.value.cancelling) return;
  poolDeleteLoading.value = true;
  error.value = null;
  try {
    const result = await getAllFilteredPoolMembers();
    if (!result.items.length) {
      ElMessage.info("当前筛选没有匹配的邮箱");
      return;
    }
    if (result.items.length > 1000) {
      throw Object.assign(new Error("当前筛选超过单次 Apple 批量删除上限 1000 个，请缩小筛选范围后再操作。"), { code: "BATCH_DELETE_TOO_LARGE" });
    }
    poolDeletePreview.value = {
      filters: poolDeleteFilterSummary(),
      total: result.items.length,
      ids: result.items.map((row) => row.alias_id),
      items: result.items.slice(0, 50),
    };
    poolDeleteConfirmOpen.value = true;
  } catch (e) {
    if (e.name !== "AbortError") error.value = e;
  } finally {
    poolDeleteLoading.value = false;
  }
}
async function submitPoolDelete() {
  if (poolDeleteConfirming.value || !poolDeletePreview.value.ids.length) return;
  poolDeleteConfirming.value = true;
  poolDeleteConfirmOpen.value = false;
  poolDeleteProgressOpen.value = true;
  error.value = null;
  try {
    await poolDeletionController.submit(poolDeletePreview.value.ids, auth.state.csrfToken);
  } catch (e) {
    error.value = e;
    poolDeleteProgressOpen.value = false;
  } finally {
    poolDeleteConfirming.value = false;
  }
}
async function cancelPoolDelete() {
  if (!isAliasDeletionJobActive(poolDeletionJob.value)) return;
  await poolDeletionController.cancel(auth.state.csrfToken);
}
async function loadLeases(signal) {
  const ticket = ++leaseReadTicket;
  await attempt(async () => {
    const result = await request(
      "/pool/leases?" +
        query({
          state: leaseState.value,
          purpose: leasePurpose.value,
          client_id: clientFilter.value,
          limit: 50,
          offset: (leasePage.value - 1) * 50,
        }), "GET", undefined, signal instanceof AbortSignal ? signal : undefined,
    );
    if (ticket === leaseReadTicket) leases.value = result;
  });
}
async function refresh(forceClients = false) {
  const ticket = ++poolRefreshTicket;
  poolRefreshController?.abort();
  const controller = new AbortController();
  poolRefreshController = controller;
  const activeTab = tab.value;
  loading.value = true;
  error.value = null;
  try {
    await Promise.all([
      request("/pool/settings", "GET", undefined, controller.signal).then((result) => {
        if (ticket === poolRefreshTicket) immediateUsed.value = result.immediate_used;
      }),
      (forceClients || !clientsLoaded.value
        ? loadClients(forceClients).then((result) => {
            if (ticket === poolRefreshTicket) {
              clients.value = result;
              clientsLoaded.value = true;
            }
          })
        : Promise.resolve()),
      activeTab === "members" ? Promise.all([
        loadMembers(controller.signal),
        getAllAccounts({ signal: controller.signal }).then((result) => {
          if (ticket === poolRefreshTicket) mainAccounts.value = result;
        }),
      ]) :
        activeTab === "leases" ? loadLeases(controller.signal) :
          activeTab === "accounts" ? request("/pool/accounts", "GET", undefined, controller.signal).then((result) => {
            if (ticket === poolRefreshTicket) accounts.value = result;
          }) : Promise.resolve(),
    ]);
  } catch (e) {
    if (ticket === poolRefreshTicket && e.name !== "AbortError") error.value = e;
  } finally {
    if (ticket === poolRefreshTicket) loading.value = false;
  }
}
function loadClients(force = false) {
  if (clientsInFlight && !force) return clientsInFlight;
  clientsLoading.value = true;
  const promise = request("/pool/clients", "GET").finally(() => {
    if (clientsInFlight === promise) {
      clientsInFlight = undefined;
      clientsLoading.value = false;
    }
  });
  clientsInFlight = promise;
  return promise;
}
function refreshTab() {
  return refresh(false);
}
onBeforeUnmount(() => {
  ++poolRefreshTicket;
  ++memberReadTicket;
  ++leaseReadTicket;
  poolRefreshController?.abort();
  poolDeletionController.stop();
});
async function mutate(fn) {
  if (busy.value) return;
  busy.value = true;
  error.value = null;
  try {
    await fn();
    await refresh(true);
  } catch (e) {
    if (e !== "cancel" && e !== "close") error.value = e;
  } finally {
    busy.value = false;
  }
}
function reveal(title, value) {
  secretTitle.value = title;
  secret.value =
    typeof value === "string" ? value : JSON.stringify(value, null, 2);
  secretOpen.value = true;
}
async function createClient() {
  await mutate(async () => {
    const result = await request("/pool/clients", "POST", {
      project: projectName.value,
    });
    reveal("项目 Key（本次显示）", result.api_key);
    projectName.value = "";
  });
}
async function updateClient(row, rotate) {
  await mutate(async () => {
    if (rotate)
      await ElMessageBox.confirm(
        "轮换后旧项目 Key 立即失效，领取记录保留。",
        "轮换项目 Key",
        { type: "warning" },
      );
    const result = await request("/pool/clients/" + row.id, "PATCH", {
      enabled: rotate ? row.enabled : !row.enabled,
      rotate,
    });
    if (rotate) reveal("新的项目 Key", result.api_key);
  });
}
async function setMember(row) {
  await mutate(() =>
    request("/pool/members/" + row.alias_id, "PATCH", {
      state: row.state === "paused" ? "available" : "paused",
    }),
  );
}
async function actLease(row, action) {
  await mutate(async () => {
    await ElMessageBox.confirm(
      action === "release"
        ? "释放后会轮换邮箱凭据，并重新进入空闲库存。"
        : "确认后该邮箱保留给当前项目。",
      action === "release" ? "释放邮箱" : "确认使用",
    );
    await request("/pool/leases/" + row.id + "/" + action, "POST");
  });
}
async function saveAccount(row) {
  await mutate(async () => {
    await request("/pool/accounts/" + row.account_id, "PUT", {
      auto_enroll: row.auto_enroll,
      target: row.target,
    });
    ElMessage.success("库存设置已保存");
  });
}
async function saveImmediateUsed(enabled) {
  await mutate(async () => {
    await ElMessageBox.confirm(
      enabled ? "新领取将立即标记为已使用，只有主动 release 才会回池。" : "新领取将改为有期限的 leased，需 commit 才会长期占用。",
      "切换领取模式",
      { type: "warning" },
    );
    await request("/pool/settings", "PUT", { immediate_used: enabled });
    ElMessage.success("领取模式已保存");
  });
}
async function toggleAutoCreate(row, enabled) {
  await mutate(async () => {
    await request("/pool/accounts/" + row.account_id, "PUT", {
      auto_enroll: row.auto_enroll,
      target: row.target,
    });
    await request(
      "/accounts/" + row.account_id + "/aliases/auto-create",
      "PUT",
      { enabled },
    );
    row.auto_create = enabled;
  });
}
function beginEnrollRead() {
  enrollDrag.stop();
  enrollAbortController?.abort();
  const ticket = ++enrollTicket;
  enrollAbortController = new AbortController();
  return { signal: enrollAbortController.signal, current: () => enrollOpen.value && ticket === enrollTicket };
}

async function restoreEnrollSelection() {
  restoringEnrollSelection = true;
  try {
    await nextTick();
    const ids = new Set(selected.value);
    for (const row of enrollItems.value) enrollTable.value?.toggleRowSelection(row, ids.has(row.id));
    await nextTick();
  } finally {
    restoringEnrollSelection = false;
  }
}

async function loadEnroll({ includeAccounts = false } = {}) {
  const scope = beginEnrollRead();
  enrollLoading.value = true;
  error.value = null;
  try {
    const [result, accountList] = await Promise.all([
      getAliasPage("", { query: enrollSearch.value, enabled: true, limit: enrollPageSize, offset: (enrollPage.value - 1) * enrollPageSize, signal: scope.signal }),
      includeAccounts ? getAllAccounts({ signal: scope.signal }) : Promise.resolve(null),
    ]);
    if (!scope.current()) return;
    if (accountList) enabledAccountIds.value = new Set(accountList.filter((a) => a.enabled).map((a) => a.id));
    restoringEnrollSelection = true;
    enrollItems.value = result.items;
    enrollTotal.value = result.total;
    const stale = new Set(result.items.filter((row) => !isSelectable(row)).map((row) => row.id));
    selected.value = selected.value.filter((id) => !stale.has(id));
    await restoreEnrollSelection();
  } catch (e) {
    if (scope.current() && e.name !== "AbortError") error.value = e;
  } finally {
    if (scope.current()) enrollLoading.value = false;
  }
}

async function openEnroll() {
  if (busy.value || enrollOpen.value) return;
  enrollOpen.value = true;
  enrollPage.value = 1;
  enrollDraft.value = enrollSearch.value = enrollProgress.value = "";
  selected.value = [];
  enrollItems.value = [];
  enrollTotal.value = 0;
  enabledAccountIds.value = new Set();
  await loadEnroll({ includeAccounts: true });
}

function closeEnroll() {
  if (!enrollSubmitting.value) enrollOpen.value = false;
}

function resetEnroll() {
  enrollDrag.stop();
  ++enrollTicket;
  enrollAbortController?.abort();
  selected.value = [];
  enrollItems.value = [];
  enrollLoading.value = enrollSelectingAll.value = false;
}
watch(enrollOpen, (open) => { if (!open) resetEnroll(); }, { flush: "sync" });
onBeforeUnmount(resetEnroll);

function isSelectable(row) {
  return poolAliasSelectable(row, enabledAccountIds.value);
}

function onEnrollSelectionChange(rows) {
  if (restoringEnrollSelection || enrollLocked.value || !enrollOpen.value) return;
  selected.value = mergePoolPageSelection(selected.value, enrollItems.value, rows.filter(isSelectable));
}

async function selectCurrentPage() {
  if (enrollLocked.value) return;
  selected.value = [...new Set([...selected.value, ...enrollItems.value.filter(isSelectable).map((row) => row.id)])];
  await restoreEnrollSelection();
}

async function clearSelection() {
  if (enrollLocked.value) return;
  enrollDrag.stop();
  selected.value = [];
  await restoreEnrollSelection();
}

async function applyEnrollSearch() {
  if (enrollLocked.value) return;
  enrollSearch.value = enrollDraft.value.trim();
  enrollPage.value = 1;
  selected.value = [];
  await loadEnroll();
}

async function selectAllResults() {
  if (enrollLocked.value) return;
  const scope = beginEnrollRead();
  enrollSelectingAll.value = true;
  error.value = null;
  try {
    const items = await getAllAliases("", { query: enrollSearch.value, enabled: true, signal: scope.signal });
    if (!scope.current()) return;
    selected.value = [...new Set(items.filter(isSelectable).map((row) => row.id))];
    await restoreEnrollSelection();
  } catch (e) {
    if (scope.current() && e.name !== "AbortError") error.value = e;
  } finally {
    if (scope.current()) enrollSelectingAll.value = false;
  }
}

async function changeEnrollPage(page) {
  if (enrollLocked.value) return;
  enrollPage.value = Math.max(1, Number(page) || 1);
  await loadEnroll();
}

async function enroll() {
  if (enrollLocked.value || !selected.value.length) return;
  enrollDrag.stop();
  const ticket = enrollTicket;
  const current = () => enrollOpen.value && ticket === enrollTicket;
  enrollSubmitting.value = true;
  error.value = null;
  try {
    await submitPoolEnrollment(selected.value,
      (ids) => request("/pool/members", "POST", { alias_ids: ids }),
      async ({ done, total, remaining }) => {
        selected.value = remaining;
        enrollProgress.value = `已完成 ${done}/${total}`;
        await restoreEnrollSelection();
      }, current);
    if (current()) {
      enrollOpen.value = false;
      await refresh(true);
    }
  } catch (e) {
    if (current()) error.value = e;
  } finally {
    enrollSubmitting.value = false;
  }
}
async function showCredentials(id) {
  await attempt(async () => {
    const result = await request("/aliases/" + id);
    const a = result.alias || result;
    reveal("邮箱完整凭据", {
      email: a.address,
      api_key: a.api_key,
      imap_password: a.imap_password,
      client_id: a.client_id,
      refresh_token: a.refresh_token,
      otp_url: new URL(a.otp_url_path, location.origin).href,
    });
  });
}
async function copySecret() {
  await attempt(async () => {
    await copyText(secret.value);
    ElMessage.success("已复制");
  });
}
const claimExample = computed(
  () =>
    `curl '${location.origin}/api/v1/pool/claim' \\\n  -H 'Authorization: Bearer PROJECT_KEY' \\\n  -H 'Content-Type: application/json' \\\n  -d '{"request_id":"job-20260912-0001","count":1,"ttl_seconds":1800,"purpose":"common"}'`,
);
const codeExample = computed(
  () =>
    `curl '${location.origin}/api/v1/pool/leases/LEASE_ID/code' \\\n  -H 'Authorization: Bearer PROJECT_KEY'`,
);
onMounted(async () => {
  await refresh(true);
  if (auth.state.username) void poolDeletionController.start();
});
</script>

<style scoped>
.pool-page {
  min-width: 0;
}
.pool-table-scroll {
  min-width: 0;
  max-width: 100%;
  overflow-x: auto;
  overscroll-behavior-x: contain;
}
.pool-table-scroll .pool-table {
  min-width: 760px;
}
.pool-toolbar {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin: 16px 0;
}
.pool-toolbar .el-select,
.pool-toolbar .el-input {
  width: 230px;
}
.pool-member-account {
  display: block;
  margin-top: 2px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}
.pool-hint {
  line-height: 1.8;
  color: var(--el-text-color-secondary);
  margin: 14px 0 20px;
}

.enroll-pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 12px;
  color: var(--text-secondary);
  font-size: 13px;
}

@media (max-width: 560px) {
  .enroll-pagination {
    align-items: flex-start;
    flex-direction: column;
  }
}
.pool-error {
  color: var(--el-color-danger);
}
.pool-cooldown { color: var(--el-color-success); }
.pool-delete-warning {
  line-height: 1.8;
  color: var(--el-color-danger);
}
.pool-delete-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  margin: 14px 0;
  color: var(--el-text-color-secondary);
}
.pool-delete-filters strong {
  color: var(--el-text-color-primary);
}
.pool-delete-count {
  font-size: 18px;
  font-weight: 600;
  margin: 14px 0;
}
.pool-delete-preview {
  max-height: 280px;
  overflow: auto;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  padding: 8px 12px;
}
.pool-delete-preview__row,
.pool-delete-waits,
.pool-delete-failures {
  display: grid;
  gap: 4px;
  line-height: 1.6;
  margin: 4px 0;
}
.pool-delete-preview__row {
  grid-template-columns: minmax(0, 1fr) minmax(150px, 0.7fr);
  gap: 12px;
}
.pool-delete-preview__row small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.pool-delete-progress__header {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 12px;
}
.pool-delete-waits,
.pool-delete-failures {
  margin-top: 14px;
  color: var(--el-text-color-secondary);
}
.el-pagination {
  margin-top: 20px;
  overflow-x: auto;
}
.pool-api {
  max-width: 880px;
  line-height: 1.8;
}
.pool-api h3 {
  margin: 24px 0 8px;
}
.pool-api pre {
  padding: 18px;
  background: var(--el-fill-color-light);
  border-radius: 8px;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font-size: 13px;
}
small {
  color: var(--el-text-color-secondary);
}
.pool-page :deep(.el-table__empty-text) {
  line-height: 1.8;
  padding: 20px 0;
  width: 80%;
}
</style>
