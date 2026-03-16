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
