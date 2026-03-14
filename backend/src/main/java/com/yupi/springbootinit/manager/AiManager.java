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
