# AI 对接改造指南：鱼聪明 SDK → 百度千帆 HTTP API

## 背景

本项目跟随鱼皮教程分期开发。后端每期从教程源码覆盖更新到 `backend/` 目录。

教程原始代码使用**鱼聪明 (YuCongMing) Java SDK** 对接 AI 能力，但该平台已不再提供智能体服务。因此需要将 AI 对接方式改为**直接通过 HTTP 调用百度千帆 OpenAI 兼容 API**，不依赖任何第三方 SDK。

**核心思路**：项目已有 `hutool-all` 依赖（HTTP 工具库），直接用它发 HTTP 请求到百度千帆 API，完全不需要新增任何依赖。

## 每次覆盖后端源码后，需要执行以下改动

每次从教程拷贝新一期后端代码后，以下 4 个文件会被还原成教程原版（鱼聪明 SDK 版本），需要重新改回来。

---

### 改动 1：`pom.xml` — 注释掉鱼聪明 SDK 依赖

找到这段：

```xml
<dependency>
    <groupId>com.yucongming</groupId>
    <artifactId>yucongming-java-sdk</artifactId>
    <version>0.0.2</version>
</dependency>
```

改成：

```xml
<!-- 已改用 HTTP 直连百度千帆 API，不再需要鱼聪明 SDK -->
<!--
<dependency>
    <groupId>com.yucongming</groupId>
    <artifactId>yucongming-java-sdk</artifactId>
    <version>0.0.2</version>
</dependency>
-->
```

---

### 改动 2：`application.yml` — 替换 AI 配置

文件路径：`backend/src/main/resources/application.yml`

找到文件末尾的鱼聪明配置：

```yaml
# 鱼聪明 AI 配置（https://yucongming.com/）
yuapi:
  client:
    access-key: 替换为你自己的 access-key
    secret-key: 替换为你自己的 secret-key
```

替换为：

```yaml
# AI 配置（百度千帆 OpenAI 兼容接口）
ai:
  base-url: https://qianfan.baidubce.com/v2/coding
  api-key: bce-v3/ALTAKSP-GfyMSBwqhkHGMnFG8FxLN/01e33d7629fc31c796a2387db09a5eb2f4b65b03
  model: qianfan-code-latest
```

> **如果要换其他 AI 平台**（如硅基流动），只需改这三行配置，代码不用动：
> ```yaml
> ai:
>   base-url: https://api.siliconflow.cn/v1    # 硅基流动示例
>   api-key: 你的apikey
>   model: Qwen/Qwen2.5-7B-Instruct           # 硅基流动模型示例
> ```

---

### 改动 3：`AiManager.java` — 整个文件替换

文件路径：`backend/src/main/java/com/yupi/springbootinit/manager/AiManager.java`

将整个文件内容替换为：

```java
package com.yupi.springbootinit.manager;

import cn.hutool.http.HttpRequest;
import cn.hutool.json.JSONArray;
import cn.hutool.json.JSONObject;
import cn.hutool.json.JSONUtil;
import com.yupi.springbootinit.common.ErrorCode;
import com.yupi.springbootinit.exception.BusinessException;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

/**
 * 用于对接 AI 平台（OpenAI 兼容接口）
 */
@Service
@Slf4j
public class AiManager {

    @Value("${ai.base-url}")
    private String baseUrl;

    @Value("${ai.api-key}")
    private String apiKey;

    @Value("${ai.model}")
    private String model;

    /**
     * AI 对话（带系统预设 prompt）
     *
     * @param systemPrompt 系统预设角色提示词
     * @param userMessage  用户输入内容
     * @return AI 生成的文本内容
     */
    public String doChat(String systemPrompt, String userMessage) {
        JSONArray messages = new JSONArray();

        JSONObject systemMsg = new JSONObject();
        systemMsg.set("role", "system");
        systemMsg.set("content", systemPrompt);
        messages.add(systemMsg);

        JSONObject userMsg = new JSONObject();
        userMsg.set("role", "user");
        userMsg.set("content", userMessage);
        messages.add(userMsg);

        JSONObject requestBody = new JSONObject();
        requestBody.set("model", model);
        requestBody.set("messages", messages);
        requestBody.set("temperature", 0.2);

        String url = baseUrl + "/chat/completions";
        log.info("AI 请求地址: {}", url);

        String responseStr;
        try {
            responseStr = HttpRequest.post(url)
                    .header("Authorization", "Bearer " + apiKey)
                    .header("Content-Type", "application/json")
                    .body(requestBody.toString())
                    .timeout(120000)
                    .execute()
                    .body();
        } catch (Exception e) {
            log.error("AI 请求失败", e);
            throw new BusinessException(ErrorCode.SYSTEM_ERROR, "AI 接口调用失败: " + e.getMessage());
        }

        log.info("AI 原始响应: {}", responseStr);

        JSONObject responseJson = JSONUtil.parseObj(responseStr);

        if (responseJson.containsKey("error")) {
            String errorMsg = responseJson.getJSONObject("error").getStr("message", "未知错误");
            log.error("AI 返回错误: {}", errorMsg);
            throw new BusinessException(ErrorCode.SYSTEM_ERROR, "AI 返回错误: " + errorMsg);
        }

        JSONArray choices = responseJson.getJSONArray("choices");
        if (choices == null || choices.isEmpty()) {
            throw new BusinessException(ErrorCode.SYSTEM_ERROR, "AI 响应为空");
        }

        return choices.getJSONObject(0).getJSONObject("message").getStr("content");
    }
}
```

**和原版的区别**：
- 原版：注入 `YuCongMingClient`（SDK 客户端），调用 `yuCongMingClient.doChat(modelId, message)`
- 新版：从 `application.yml` 读取 `base-url`/`api-key`/`model`，用 Hutool HTTP 发送 POST 请求到 OpenAI 兼容接口
- 方法签名从 `doChat(long modelId, String message)` 改为 `doChat(String systemPrompt, String userMessage)`

---

### 改动 4：`ChartController.java` — 修改 AI 调用部分

文件路径：`backend/src/main/java/com/yupi/springbootinit/controller/ChartController.java`

找到 `genChartByAi` 方法（`@PostMapping("/gen")`）中的 AI 调用部分。

#### 找到这段（原版）：

```java
        // 无需写 prompt，直接调用现有模型，https://www.yucongming.com，公众号搜【鱼聪明AI】
//        final String prompt = "你是一个数据分析师和前端开发专家，接下来我会按照以下固定格式给你提供内容：\n" +
//                "分析需求：\n" +
//                "{数据分析的需求或者目标}\n" +
//                "原始数据：\n" +
//                "{csv格式的原始数据，用,作为分隔符}\n" +
//                "请根据这两部分内容，按照以下指定格式生成内容（此外不要输出任何多余的开头、结尾、注释）\n" +
//                "【【【【【\n" +
//                "{前端 Echarts V5 的 option 配置对象js代码，合理地将数据进行可视化，不要生成任何多余的内容，比如注释}\n" +
//                "【【【【【\n" +
//                "{明确的数据分析结论、越详细越好，不要生成多余的注释}";
        long biModelId = 1659171950288818178L;
        // 分析需求：
        // 分析网站用户的增长情况
        // 原始数据：
        // 日期,用户数
        // 1号,10
        // 2号,20
        // 3号,30

        // 构造用户输入
        StringBuilder userInput = new StringBuilder();
        userInput.append("分析需求：").append("\n");

        // 拼接分析目标
        String userGoal = goal;
        if (StringUtils.isNotBlank(chartType)) {
            userGoal += "，请使用" + chartType;
        }
        userInput.append(userGoal).append("\n");
        userInput.append("原始数据：").append("\n");
        // 压缩后的数据
        String csvData = ExcelUtils.excelToCsv(multipartFile);
        userInput.append(csvData).append("\n");

        String result = aiManager.doChat(biModelId, userInput.toString());
        String[] splits = result.split("【【【【【");
        if (splits.length < 3) {
            throw new BusinessException(ErrorCode.SYSTEM_ERROR, "AI 生成错误");
        }
        String genChart = splits[1].trim();
        String genResult = splits[2].trim();
```

#### 替换为：

```java
        final String systemPrompt = "你是一个数据分析师和前端开发专家，接下来我会按照以下固定格式给你提供内容：\n" +
                "分析需求：\n" +
                "{数据分析的需求或者目标}\n" +
                "原始数据：\n" +
                "{csv格式的原始数据，用,作为分隔符}\n" +
                "请严格按照下面的输出格式生成结果，且不得添加任何多余内容（例如无关文字、注释、代码块标记或反引号）：\n" +
                "【【【【\n" +
                "{生成 Echarts V5 的 option 配置对象 JSON 代码，要求为合法 JSON 格式且不含任何额外内容}\n" +
                "【【【【\n" +
                "结论：{提供对数据的详细分析结论，内容应尽可能准确、详细，不允许添加其他无关文字或注释}";

        // 构造用户输入
        StringBuilder userInput = new StringBuilder();
        userInput.append("分析需求：").append("\n");

        String userGoal = goal;
        if (StringUtils.isNotBlank(chartType)) {
            userGoal += "，请使用" + chartType;
        }
        userInput.append(userGoal).append("\n");
        userInput.append("原始数据：").append("\n");
        String csvData = ExcelUtils.excelToCsv(multipartFile);
        userInput.append(csvData).append("\n");

        String result = aiManager.doChat(systemPrompt, userInput.toString());
        String[] splits = result.split("【【【【");
        if (splits.length < 3) {
            throw new BusinessException(ErrorCode.SYSTEM_ERROR, "AI 生成错误");
        }
        String genChart = splits[1].trim();
        String genResult = splits[2].trim();
```

**关键变化**：
1. 去掉 `long biModelId = ...`，改为定义 `systemPrompt` 字符串（AI 角色预设 + 输出格式要求）
2. 调用方式从 `aiManager.doChat(biModelId, userInput)` 改为 `aiManager.doChat(systemPrompt, userInput)`
3. 分隔符从 `【【【【【`（5个）改为 `【【【【`（4个），和 prompt 中的分隔符一致

---

### 改动 5（可选）：`AiManagerTest.java` — 更新测试类

文件路径：`backend/src/test/java/com/yupi/springbootinit/manager/AiManagerTest.java`

将整个文件内容替换为：

```java
package com.yupi.springbootinit.manager;

import org.junit.jupiter.api.Test;
import org.springframework.boot.test.context.SpringBootTest;

import javax.annotation.Resource;

@SpringBootTest
class AiManagerTest {

    @Resource
    private AiManager aiManager;

    @Test
    void doChat() {
        String systemPrompt = "你是一个数据分析师和前端开发专家，接下来我会按照以下固定格式给你提供内容：\n" +
                "分析需求：\n" +
                "{数据分析的需求或者目标}\n" +
                "原始数据：\n" +
                "{csv格式的原始数据，用,作为分隔符}\n" +
                "请严格按照下面的输出格式生成结果，且不得添加任何多余内容（例如无关文字、注释、代码块标记或反引号）：\n" +
                "【【【【\n" +
                "{生成 Echarts V5 的 option 配置对象 JSON 代码，要求为合法 JSON 格式且不含任何额外内容}\n" +
                "【【【【\n" +
                "结论：{提供对数据的详细分析结论，内容应尽可能准确、详细，不允许添加其他无关文字或注释}";

        String userMessage = "分析需求：\n" +
                "分析网站用户的增长情况，请使用柱状图\n" +
                "原始数据：\n" +
                "日期,用户数\n" +
                "1号,10\n" +
                "2号,20\n" +
                "3号,30\n";

        String answer = aiManager.doChat(systemPrompt, userMessage);
        System.out.println(answer);
    }
}
```

---

## 验证方法

改完后运行单元测试验证 AI 接口是否通：

```bash
cd backend
mvn test -Dtest=AiManagerTest#doChat -pl .
```

如果控制台输出中能看到包含 `【【【【` 分隔的 ECharts JSON 和分析结论，说明改造成功。

---

## 数据库字段改动与实体类同步

### 背景说明

如果你在 `backend/sql/create_table.sql` 中新增了数据库字段（例如为了支持异步任务添加 `status` 和 `execMessage` 字段），需要完成以下 3 个步骤才能让这些字段真正生效：

### 步骤 1：执行数据库迁移

在 Cursor 的 Database 插件中连接到数据库后：

**如果是全新数据库**：
- 直接执行整个 `create_table.sql` 文件

**如果数据库已有数据**（推荐方式）：
- 使用 ALTER TABLE 语句添加新字段，避免删除现有数据
- 例如添加 `status` 和 `execMessage` 字段：

```sql
ALTER TABLE chart 
ADD COLUMN status varchar(128) NOT NULL DEFAULT 'wait' COMMENT 'wait,running,succeed,failed' AFTER genResult;

ALTER TABLE chart 
ADD COLUMN execMessage text NULL COMMENT '执行信息' AFTER status;
```

**注意**：
- Database 插件已经选择了数据库，不需要在查询中写 `USE yubi;`
- 如果报错 "Duplicate column name"，说明字段已存在，跳过此步骤
- 可以用 `DESCRIBE chart;` 查看表的当前结构

### 步骤 2：更新 Java 实体类

文件路径：`backend/src/main/java/com/yupi/springbootinit/model/entity/Chart.java`

在对应位置添加新字段的属性。例如在 `genResult` 字段后、`userId` 字段前添加：

```java
/**
 * 任务状态
 */
private String status;

/**
 * 执行信息
 */
private String execMessage;
```

**为什么必须添加**：
- Java 使用 MyBatis-Plus 做 ORM（对象关系映射）
- 数据库的列 ↔️ Java 类的属性，必须一一对应
- 如果实体类中没有这个属性，Java 代码就无法读取或写入数据库中的这个字段
- 即使数据库表有这个列，Java 代码也"看不到"它

### 步骤 3：在业务逻辑中使用新字段

根据需求在 Service 或 Controller 中使用新增的字段。例如：

```java
// 创建图表时设置初始状态
Chart chart = new Chart();
chart.setStatus("wait");
chartService.save(chart);

// 更新状态
chart.setStatus("running");
chartService.updateById(chart);

// 失败时记录错误信息
chart.setStatus("failed");
chart.setExecMessage("AI 调用超时");
chartService.updateById(chart);
```

---

## 后续期数注意事项

- **第 4 期（异步分析）**：如果新增了异步版的 `genChartByAiAsync` 方法，其中的 `aiManager.doChat(biModelId, ...)` 调用同样需要改成 `aiManager.doChat(systemPrompt, ...)`，分隔符同样改为 4 个 `【`
- **第 5 期（RabbitMQ）**：消息消费者中如果调用了 `aiManager.doChat`，也需要同样修改
- **通用规则**：在整个后端代码中搜索 `biModelId` 或 `1659171950288818178L`，所有出现的地方都需要按上述方式改造
