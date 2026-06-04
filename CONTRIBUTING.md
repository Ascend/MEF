# MEF贡献指南

感谢您考虑为 MEF 做出贡献！我们欢迎任何形式的贡献，包括错误修复、功能增强、文档改进等，甚至只是反馈。无论您是经验丰富的开发者还是第一次参与开源项目，您的帮助都是非常宝贵的。

您可以通过多种方式支持本项目：

- 通过[Issues](https://gitcode.com/Ascend/MEF/issues)反馈问题，详见[Issue提交指南](https://gitcode.com/Ascend/community/blob/master/docs/contributor/issue-guide.md)。
- 建议或实现新功能。
- 改进或扩展文档。
- 审查Pull Request并协助其他贡献者。
- 传播项目：在博客文章、社交媒体上分享MEF，或给仓库点个⭐。

## 贡献流程

### 代码开发

1. 先在GitCode平台点击仓库右上角"Fork"按钮，将仓库Fork到个人仓

2. 将Fork到个人仓的代码克隆到本地进行代码开发

   ```bash
   git clone https://gitcode.com/<your-username>/MEF.git
   ```

   代码开发请遵循代码规范：
   - [昇腾社区 Go 语言编程指导](https://gitcode.com/Ascend/community/blob/master/docs/contributor/Ascend-go-coding-style-guide.md)
   - [昇腾社区 Go 语言安全编程指导](https://gitcode.com/Ascend/community/blob/master/docs/contributor/Ascend-go-secure-coding-guide.md)

### 代码测试

#### 运行测试

在提交代码前，请确保所有测试通过，测试用例执行方法详见：[测试用例](README.md#测试用例)

#### 添加测试

- 为新功能添加相应的单元测试
- 确保测试覆盖主要逻辑分支
- 测试用例应该具有良好的可读性和维护性
- 测试用例应该放在源代码同目录下
- 测试用例开发请遵循测试贡献指南：[Ascend 社区开发者测试贡献指南](https://gitcode.com/Ascend/community/blob/master/docs/contributor/developer-testing-guide.md)

### 文档开发

#### 文档路径

若您的更改涉及新增、变更或删除特性或接口，请更新相关文档，用户指南文档位置：`docs/zh/`

#### 文档规范

- 使用简洁明了的中文表述
- 提供完整的示例代码
- 包含必要的截图或图表说明
- 确保链接的有效性

### 提交Pull Request流程

#### 提交前检查清单

在提交Pull Request之前，请确保：

- [ ] 代码遵循项目的编码规范
- [ ] 添加了必要的测试用例
- [ ] 所有测试通过
- [ ] 更新了相关文档
- [ ] 提交信息清晰明确
- [ ] 代码已经过自我审查

#### 提交流程

以下为提交流程指导，详细PR提交流程说明请参见：[PR提交指南](https://gitcode.com/Ascend/community/blob/master/docs/contributor/pr-guide.md)

1. **创建本地分支**

   ```bash
   git checkout -b <new_branch_name> origin/master
   ```

2. **提交更改**

   ```bash
   git add .
   git commit -m "Your commit title"
   ```

3. **推送到远程仓库**

   ```bash
   git push origin <new_branch_name>
   ```

4. **创建Pull Request**

   在GitCode上创建Pull Request，并填写：
   - 问题/功能描述
   - 修改方案描述
   - 开发自检项勾选

5. **代码审查**

   - 提交Pull Request后，您需要通知相关“负责人”（ Reviewers和Committers）进行内容审核。
   - 您需要根据反馈审核意见修改代码，并重新提交更新。此流程可能涉及多轮迭代，请保持积极响应和沟通。

   Pull Request流程中会提示相关的“负责人”，可在Pull Request流程中指定相关“负责人”。

6. **代码合并**

   Pull Request需要依次集齐如下四个标签即可完成代码合入：

    - ascend-cla/yes：CLA检查，首次开发时需要完成CLA的签署，完成后每次提交自动获得此标签。
    - ci-pipeline-passed：CI流水线，在Pull Request流程中评论`compile`触发，若CI流水线检查不通过，则需要根据提示修改后重新提交。
    - lgtm：由Reviewers提供，Reviewers审核通过后，会在Pull Request流程中评论`/lgtm`触发lgtm标签。
    - approved：由Committers提供，Committers审核通过后，会在Pull Request流程中评论`/approved`触发approved标签。

   当您的Pull Request集齐四个标签后，您的Pull Request将被合并到主干分支。

#### Pull Request最佳实践

- 保持Pull Request的大小适中，便于审查
- 一个Pull Request只解决一个问题或实现一个功能
- 及时响应审查意见
- 保持与主分支同步，及时解决冲突

## 社区准则

### 行为准则

参与本项目应遵守：[Ascend开源项目行为守则](https://gitcode.com/Ascend/community/blob/master/docs/contributor/code-of-conduct.md)

### 沟通渠道

- **Issues**：用于报告Bug、提出功能建议和讨论技术问题
- **Pull Requests**：用于代码审查和讨论具体实现

## 许可证

通过向本项目贡献代码，您同意您的贡献将按照项目的许可证进行授权，详见[LICENSE](LICENSE.md)文件。

## 致谢

感谢您为MEF做出的贡献。您的努力使这个项目变得更加强大和用户友好。期待您的参与！

---

如有任何疑问或需要帮助，请随时在[Issues](https://gitcode.com/Ascend/MEF/issues)中提问，或通过其他社区渠道与我们联系。
