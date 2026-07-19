# Changelog

All notable changes to this project will be documented in this file.

## [0.4.2] - 2026-07-19

### Bug Fixes

- Add improved resilience, support for auto-naming and various bug fixes([e76d40b](https://github.com/newstack-cloud/bluelink-provider-aws/commit/e76d40bb252c4e7bcb32bf60365493583ff77774))
## [0.4.1] - 2026-07-17

### Bug Fixes

- Drop patterns not compatible with re2 engine([70d74ae](https://github.com/newstack-cloud/bluelink-provider-aws/commit/70d74ae7d8e8c787a1a7e38a3e278a5dc720b886))
- Correct overlays misreading reference nodes as empty lists([13ffee3](https://github.com/newstack-cloud/bluelink-provider-aws/commit/13ffee3a45798377e229d2a84b010dfa4284849f))

### Dependencies

- Bump blueprint lib to 0.51.1([1cc63b4](https://github.com/newstack-cloud/bluelink-provider-aws/commit/1cc63b4e5a0dc0650913b61f2f9f50b33f111c8a))
## [0.4.0] - 2026-07-17

### Bug Fixes

- **links:** Validate redis auth token before ModifyReplicationGroup([b8773c8](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b8773c85591a77e29fff224734dde5fbc786b29d))
- **links:** Default redis auth token strategy to ROTATE, reject initial SET([a24d766](https://github.com/newstack-cloud/bluelink-provider-aws/commit/a24d766ab64687d68f27ffbcab43bee3fee8f8cc))
- **links:** Fall back to ROTATE when SET is selected on first configuration([cdd17ab](https://github.com/newstack-cloud/bluelink-provider-aws/commit/cdd17aba6d0d75fa620d74dbb89ba851c09b67a8))
- **links:** Clear function response types when reportBatchItemFailures is false([34693c0](https://github.com/newstack-cloud/bluelink-provider-aws/commit/34693c0d1ecd4a95554178b62eac62bdc0be6098))
- **links:** Clear esm filter criteria when filter annotations are removed([b46569c](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b46569c7cb1a7552fa546ae0f83819cc2439f567))

### Features

- Add region to ssm parameter spec([4f6d654](https://github.com/newstack-cloud/bluelink-provider-aws/commit/4f6d6548a4265c2adaac1c1ff553352b0dd1d08d))
- Add ssm parameter tree and lambda param tree link([1136c4a](https://github.com/newstack-cloud/bluelink-provider-aws/commit/1136c4a41b997d5504818ac2493e2034d48da432))
- **links:** Dynamodb stream reportBatchItemFailures on lambda esm link([23b97fa](https://github.com/newstack-cloud/bluelink-provider-aws/commit/23b97faaafc92f50e444849e506ded10cd5c48d0))
- **links:** Elasticache replicationGroup::secret auth-token link([a3cd2eb](https://github.com/newstack-cloud/bluelink-provider-aws/commit/a3cd2eb03d31b3078dcbbb5a9a3535c7ea9d474a))

### Refactoring

- **links:** Trim comments on dynamodb and elasticache links([bc6495f](https://github.com/newstack-cloud/bluelink-provider-aws/commit/bc6495fb97fc8873ef16e52faaf9b351dd92fcd2))
## [0.3.1] - 2026-07-10

### Bug Fixes

- Add per-tier subnet outputs and add subnet group validation([01742ad](https://github.com/newstack-cloud/bluelink-provider-aws/commit/01742ad267d9239d7147e2af9e6c171e81a19cab))
## [0.3.0] - 2026-07-10

### Features

- Add support for scoping links to ssm parameter path prefixes([e497bbf](https://github.com/newstack-cloud/bluelink-provider-aws/commit/e497bbf56751555c4271c42415ad7eeac6098eb7))
## [0.2.0] - 2026-07-09

### Bug Fixes

- Correct field path in iam policy link utils([667c137](https://github.com/newstack-cloud/bluelink-provider-aws/commit/667c13758e17d87ed3d1115b8459553674c68761))
- Make sure flex vpc is registered as a stablised dependency of other resource types([d17c62a](https://github.com/newstack-cloud/bluelink-provider-aws/commit/d17c62a74f0f22f78485c3a5c56a5b2761e46e8c))
- Add fix for qualifier bug in lambda function data source([f1d4cf1](https://github.com/newstack-cloud/bluelink-provider-aws/commit/f1d4cf134eed3a08d8d66dece026c5f9f09ae868))

### Dependencies

- Bump up bluelink libs([12c34f6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/12c34f6e98b71f46359c8e818d7dd55173a1ef46))

### Features

- Add cardinality constraint to function signing config link([4aeccf8](https://github.com/newstack-cloud/bluelink-provider-aws/commit/4aeccf81306c983fde4fc7b8efa29adcc5797c2b))
- Add cardinality constraint to source queue to dlq link([807d799](https://github.com/newstack-cloud/bluelink-provider-aws/commit/807d7999ffaeb6dba9b879f9f1ec0c8d88e01681))
- Update example docs for iam resources([3b76b8a](https://github.com/newstack-cloud/bluelink-provider-aws/commit/3b76b8a4c6247e3b1ebc24a6d31ec3804648ff98))
- Add support for event bridge and wire up missing resources([faa25a3](https://github.com/newstack-cloud/bluelink-provider-aws/commit/faa25a366a6d3a98d6c994358ffdcda6f4bbaf64))
- Add initial generated output from cloud control resource and ds generation([0d9a918](https://github.com/newstack-cloud/bluelink-provider-aws/commit/0d9a9185f2f18792dc0535ca57c2e3604a9a563a))
- Add cloud control resource and data source implementation([4da2194](https://github.com/newstack-cloud/bluelink-provider-aws/commit/4da219417bc4a731a451823d60d48e3552666e08))
- Add code gen tool for cloud control backed resources and data sources([bf732df](https://github.com/newstack-cloud/bluelink-provider-aws/commit/bf732df5b1a550cc166bca8696ace5bb640906c6))
- Add and update helpers for link-managed resources and permissions([8efbe0d](https://github.com/newstack-cloud/bluelink-provider-aws/commit/8efbe0dac72a149f6a6803d816dd2a11652861fd))
- Wire up cloud control resources and new links in the provider([2c501c7](https://github.com/newstack-cloud/bluelink-provider-aws/commit/2c501c7a903549bef40c70f7919c06bce63a10d4))
- Add link implementation for event bridge with sqs and lambda([f544029](https://github.com/newstack-cloud/bluelink-provider-aws/commit/f5440290ab4123148ccbd701d38448831cd2caf1))
- Update permission allocation logic for lambda dynamodb links along with other improvements([8f841b3](https://github.com/newstack-cloud/bluelink-provider-aws/commit/8f841b31c3bfd6b14bab12b886ba4f77fd074053))
- Add network link activation helper and vpc lambda function link([0bf0158](https://github.com/newstack-cloud/bluelink-provider-aws/commit/0bf0158a776a8d164a88c2425c8da93b4f7e6ca0))
- Add support for sns resources and links([0352279](https://github.com/newstack-cloud/bluelink-provider-aws/commit/03522792af8290fdeb115d8919abefe7734623ee))
- Update ddb stream lambda link to use new index-based filter keys([1cebdee](https://github.com/newstack-cloud/bluelink-provider-aws/commit/1cebdeefa6bc66cc3524222277a671d76b695e2e))
- Add kinesis streams along with stream/queue to lambda links([bdbc4c8](https://github.com/newstack-cloud/bluelink-provider-aws/commit/bdbc4c8f6514f0e2fddd51eed480baecf120142d))
- Add bucket resource and links for bucket access and notifications([fee46ed](https://github.com/newstack-cloud/bluelink-provider-aws/commit/fee46edad31471d14125e9858ca824b870cebed8))
- Add ssm, kms and secrets manager support with links from lambda functions([adca847](https://github.com/newstack-cloud/bluelink-provider-aws/commit/adca847c8f62131620daebccd8607cea00f16258))
- Add first phase of support for rds focused on rds proxy and networking([f5e18e7](https://github.com/newstack-cloud/bluelink-provider-aws/commit/f5e18e71cd1e90d9d1d0914b3b596db4fd246926))
- Add support for aurora cluster with function cluster link support([0b26c16](https://github.com/newstack-cloud/bluelink-provider-aws/commit/0b26c16df6a61fcb8e0ecbaf340dddecfefcd5ac))
- Add support for elasticache and add lambda cache links([db3279c](https://github.com/newstack-cloud/bluelink-provider-aws/commit/db3279ca453716d6c3c601e5145225c0dc98f07b))
- Add api gateway v2 support with lambda links([6f1509b](https://github.com/newstack-cloud/bluelink-provider-aws/commit/6f1509b10930db5e75834a0bb5312a8ccf3e4e1e))
- Add lambda function to sqs queue link implementation([00cca31](https://github.com/newstack-cloud/bluelink-provider-aws/commit/00cca310b2b92357fb59c0c57539f7b487943462))

### Refactoring

- Clean up old manual implementation, update docs and test script([fad0052](https://github.com/newstack-cloud/bluelink-provider-aws/commit/fad0052c311b760ac0c1fce5677a2a76e9120186))

### Testing

- Add e2e/integration test harness and initial test suites([a83f602](https://github.com/newstack-cloud/bluelink-provider-aws/commit/a83f6020c5be98142bd96a9a71cc17ea19613ab6))
- Add direct call integration test harness([bf707fc](https://github.com/newstack-cloud/bluelink-provider-aws/commit/bf707fc05d44658a3548d2f5240dd0bdd960de22))
- Add service mocks for unit tests([6bec4fd](https://github.com/newstack-cloud/bluelink-provider-aws/commit/6bec4fd1a239132a9a3de7669f7a71f66f5993ba))
- Add integration test for lambda layer versions([40dce2d](https://github.com/newstack-cloud/bluelink-provider-aws/commit/40dce2dea0b1b1c68adcc84fd4e83b4f00b47985))
- Add missing tests for link implementations([9983972](https://github.com/newstack-cloud/bluelink-provider-aws/commit/9983972a25694e6a65cebbf091515aeb6a11a53b))
## [0.1.1] - 2026-04-08

### Bug Fixes

- Bump for initial 0.1.1 release([2229cd4](https://github.com/newstack-cloud/bluelink-provider-aws/commit/2229cd4df43ebb2ccabc9d24997ea033099d45a9))
## [0.1.0] - 2026-04-08

### Bug Fixes

- Add missing nil checks and a full test suite for get external state([4c09ef0](https://github.com/newstack-cloud/bluelink-provider-aws/commit/4c09ef028f239237cde84c9d347b870484ab02e1))
- Add corrections to function version examples([690f31a](https://github.com/newstack-cloud/bluelink-provider-aws/commit/690f31a44576c97a959da59c6c3ee517166c28b4))
- Correct formatting in layer version get external state file([9bd563e](https://github.com/newstack-cloud/bluelink-provider-aws/commit/9bd563e1f0186255522aead60d723f6d6020bcc7))
- Remove zipFile field from lambda layer version([6f1aba5](https://github.com/newstack-cloud/bluelink-provider-aws/commit/6f1aba58cbd08296faca83b1aa4d03e07eba3879))
- Correct examples for lambda layer version resource([dc57ae9](https://github.com/newstack-cloud/bluelink-provider-aws/commit/dc57ae94ebb6bdd793ff20ee475e49348082a100))
- Ensure iam resources are registered with the plugin provider([505f2b6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/505f2b68b49025e941037f6a3cd78feee864398a))
- Add missing computed fields from oidc provider update response([f542b5d](https://github.com/newstack-cloud/bluelink-provider-aws/commit/f542b5d2c0640e769cc509d3adedcb08617a9cca))
- Add various fixes for iam and integrate limited support for provenance tagging([0f55aef](https://github.com/newstack-cloud/bluelink-provider-aws/commit/0f55aefcccfcba6ecbebeb7a73de886cf4d27176))
- Add various fixes to sqs and support for provenance tagging([181b43e](https://github.com/newstack-cloud/bluelink-provider-aws/commit/181b43eb31d284e9774e9fe9e49aabba74f680a5))
- Bump initial version to force 0.1.1 release([0351a24](https://github.com/newstack-cloud/bluelink-provider-aws/commit/0351a24425c702accb47db54382f0bb43e4975cd))
- Force version bump to 0.1.1([449276c](https://github.com/newstack-cloud/bluelink-provider-aws/commit/449276c5d0dd26e15aa8ef1f2f5861dba1ffe8b5))
- Force next release as 0.2.0([13efd52](https://github.com/newstack-cloud/bluelink-provider-aws/commit/13efd5225d953547dfd2934285574abd9160bbe0))
- Update docs to get first proper release([c00a0fc](https://github.com/newstack-cloud/bluelink-provider-aws/commit/c00a0fc32173d77f588caa17faba568018d5ab99))
- Force bump to 0.1.1([6aab365](https://github.com/newstack-cloud/bluelink-provider-aws/commit/6aab365bb343c65fcd5dd9bd3046d466d599593d))

### Features

- Add implementation of the lambda function resource([7a519d9](https://github.com/newstack-cloud/bluelink-provider-aws/commit/7a519d9d5a67d07156a7e496ab3b2c5e1bca55ad))
- Add implementation of lambda function version resource([11c8efb](https://github.com/newstack-cloud/bluelink-provider-aws/commit/11c8efb8fe7c28f132d594df98613613cf27dc46))
- Add lambda alias resource implementation([21d825f](https://github.com/newstack-cloud/bluelink-provider-aws/commit/21d825f75d8233781067824989c54045db0946d4))
- Add lambda code signing config resource implementation([10ae1ca](https://github.com/newstack-cloud/bluelink-provider-aws/commit/10ae1ca04a41c32ebb821557eb6919c0a7074c89))
- Add alias resource implementation([d8c4188](https://github.com/newstack-cloud/bluelink-provider-aws/commit/d8c418825805f778391df6a038e950c02b9baaf4))
- Add function version resource implementation([6a531a7](https://github.com/newstack-cloud/bluelink-provider-aws/commit/6a531a725e96817ff043fb4f62a184db8e7d5a98))
- Add code signing config resource implementation([60a2f02](https://github.com/newstack-cloud/bluelink-provider-aws/commit/60a2f02cad1acece43dcbcd4ed93e692f8a8639d))
- Add event source mapping resource implementation([903c912](https://github.com/newstack-cloud/bluelink-provider-aws/commit/903c91292f821ca31215f3b57088fbafb44d6cf2))
- Add alias data source implementation([365b8b7](https://github.com/newstack-cloud/bluelink-provider-aws/commit/365b8b7a90c2f6f3e7c0c4781132958dc89284dd))
- Add code signing config data source implementation([97944c2](https://github.com/newstack-cloud/bluelink-provider-aws/commit/97944c2c8f5b660cf45b34280c8c3949f33eb094))
- Add function url resource implementation([2755b27](https://github.com/newstack-cloud/bluelink-provider-aws/commit/2755b27ba8abc7612d4a0bbdf1c36024f9355c08))
- Add implementation of lambda layer version resource([19e6a6a](https://github.com/newstack-cloud/bluelink-provider-aws/commit/19e6a6a56ebb10887f601abae63debb1253472b6))
- Add lambda event invoke config resource implementation([c00add2](https://github.com/newstack-cloud/bluelink-provider-aws/commit/c00add26156940f5e2f8ba86ee84febbbe99ea0e))
- Add lambda layer version permission resource implementation([6dede22](https://github.com/newstack-cloud/bluelink-provider-aws/commit/6dede228de7160098649914dc70eb425f8cc95f4))
- Add lambda function url data source implementation([9a34b1e](https://github.com/newstack-cloud/bluelink-provider-aws/commit/9a34b1e7c58186aad5fc234f114b15f7381ebe25))
- Add lambda layer version data source implementation([b49be69](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b49be69f833969592bdc7ccd47fc37990c6efa8c))
- Add lambda function to code signing config link implementation([1990c78](https://github.com/newstack-cloud/bluelink-provider-aws/commit/1990c7885ab8562bc25c7d24ab876e5cac0e1dda))
- Add iam role resource implementation([3e0fb0e](https://github.com/newstack-cloud/bluelink-provider-aws/commit/3e0fb0ecd39706404d4ac9558f4624873289e764))
- Update function to csc link to return resource data mappings([073dead](https://github.com/newstack-cloud/bluelink-provider-aws/commit/073dead45a7b9339d83f1805808acbb02142fb87))
- Add missing tag and policy diff behaviour for roles([b3260a6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b3260a6af5c16a297641910832ea3d262f152e53))
- Add missing perm boundary updates and tag sorting([89e7ac6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/89e7ac682bb3c08e2496dd903d84b5740e50ca0b))
- Add iam group resource implementation([a36e60b](https://github.com/newstack-cloud/bluelink-provider-aws/commit/a36e60b60157901ad31f2439ff89c9b4b57f1b49))
- Add iam access key resource implementation([151dc7d](https://github.com/newstack-cloud/bluelink-provider-aws/commit/151dc7d58173a136691e21b2435700ec1272b3ef))
- Add iam instance profile resource implementation([b3158a1](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b3158a178a1104c00fce9372faa1eb14c2c77d12))
- Add iam managed policy resource([f76fedc](https://github.com/newstack-cloud/bluelink-provider-aws/commit/f76fedc1aa9b43967c1bdca45dfb6beb370c620f))
- Add iam oidc provider resource([dd4a8b9](https://github.com/newstack-cloud/bluelink-provider-aws/commit/dd4a8b9db1b437b7eba1203b81e4b838dd5042bf))
- Add iam saml provider resource implementation([a9936cf](https://github.com/newstack-cloud/bluelink-provider-aws/commit/a9936cf4b3bd331f289daa1bbc60c98a484a1526))
- Add iam server certificate resource([51f95a8](https://github.com/newstack-cloud/bluelink-provider-aws/commit/51f95a861a3515b67ed0ee3af688062f3ae69f74))
- Add flex vpc resource implementation([b2feed4](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b2feed4882156ff1d073ba5dd58fbcb5b402c665))
- Integrate provenance tagging into flex vpc([dfd1ac8](https://github.com/newstack-cloud/bluelink-provider-aws/commit/dfd1ac880d07ff30b6b610cbfd23e5f85f2b653c))
- Integrate support for provenance tagging in lambda resources([cb16e71](https://github.com/newstack-cloud/bluelink-provider-aws/commit/cb16e710820ab392e0aefd013a231fc23f510e63))
- Add utils for bluelink tags and data conversions([8527aef](https://github.com/newstack-cloud/bluelink-provider-aws/commit/8527aef0ea514a08bac3fb1927cbfdac4f5729d3))
- Register new resources with the provider([c6730d8](https://github.com/newstack-cloud/bluelink-provider-aws/commit/c6730d8404acbc1bedebbbf609a9afebab3f6ce4))
- Add version file to bundle plugin version in code at build time([603a93d](https://github.com/newstack-cloud/bluelink-provider-aws/commit/603a93da9a543a5913f9a8429d4f5cb65bf8db57))
- Bundle build-time version into plugin metadata([b4c72f6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b4c72f6fcb1261da20268ae20db2a6bbd5b9b7f8))
- Add lambda function ddb table links and inter-service link structure([42d7a21](https://github.com/newstack-cloud/bluelink-provider-aws/commit/42d7a213fc2702facf866b9533cf41994eccd5a2))

### Refactoring

- Remove references to two-hundred([de97aa7](https://github.com/newstack-cloud/bluelink-provider-aws/commit/de97aa78b40fdc571b7f96e06234a128e4c5261d))
- Reduce code duplication by create a shared tags schema function([6d0247b](https://github.com/newstack-cloud/bluelink-provider-aws/commit/6d0247b4575c55c0f9d1b3035276e76e91586fcd))
- Reduce duplicate code with shared value extractors([172f0f6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/172f0f6967bdf6010ab5a6a7da5316ba4d3b978e))
- Reduce duplicate code for alias value extractors([13be60f](https://github.com/newstack-cloud/bluelink-provider-aws/commit/13be60f25dd3aa9d7d46c06bbc58849f0c3efbc1))
- Re-organise to share mocks and service interfaces between packages([4529ce6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/4529ce63dadbf52bf6ffc728d4564bd280bbea6d))
- Reduce duplicate code for cors setters([a84850c](https://github.com/newstack-cloud/bluelink-provider-aws/commit/a84850c89aa6cf59c90f7a4acef7be49c3359f1c))
- Update to work with bluelink([5a3c6a7](https://github.com/newstack-cloud/bluelink-provider-aws/commit/5a3c6a7b8cb941e500e31c8633fc45784b2e5bbb))
- Correct tag diff checking and reduce duplication([b7828cb](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b7828cb97890b15c546ce7c936cad6db5f8082f1))
- Add tags to differentiate between unit and integration tests([a996cf0](https://github.com/newstack-cloud/bluelink-provider-aws/commit/a996cf0f7b7bacc53534bcc34a9cb2f08a83a536))

### Testing

- Add tests for provider and custom config validation([f996f4e](https://github.com/newstack-cloud/bluelink-provider-aws/commit/f996f4e87800ad715fb33c98b84135d2b21490a9))
- Add test suite for destroying lambda functions([615d9a1](https://github.com/newstack-cloud/bluelink-provider-aws/commit/615d9a13a19f1f224cf4c5c26bbaaebd273172f0))
- Add tests for lambda function stabilised check([1cd37d7](https://github.com/newstack-cloud/bluelink-provider-aws/commit/1cd37d776fd4ff98fa5c170b6af03c67bf32dc97))
- Add missing tests for the event invoke configuration resource([81b7d35](https://github.com/newstack-cloud/bluelink-provider-aws/commit/81b7d35a256a8bf0a21020b8ca12f891bedb526a))
- Remove references to old string-based policy docs([ffec7c6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/ffec7c62cf3f18cdd2db7dd34f87c8ea5c859214))
- Add missing tests for updating iam user resource([d346454](https://github.com/newstack-cloud/bluelink-provider-aws/commit/d346454f28e240fd6480d91b14900e2ed1c5a7f5))
- Add test suites for iam group resource([b09945b](https://github.com/newstack-cloud/bluelink-provider-aws/commit/b09945b320ebfec3a2e2a66cf1d030120be2b15b))
- Add test cases for resource re-creation([90b1164](https://github.com/newstack-cloud/bluelink-provider-aws/commit/90b116441636d5c446de17374addbd6d1ad6b740))
- Add recreate test cases for lambda resources([d1d12eb](https://github.com/newstack-cloud/bluelink-provider-aws/commit/d1d12eb60f80ffbadd9f4b1670994172274e4ca6))
- Add support for integration tests([90805f6](https://github.com/newstack-cloud/bluelink-provider-aws/commit/90805f6aa5ecc8750a0ed244a322327171e2de81))
- Ensure env vars are exported in all mode and add dynamodb service mocks([958fec9](https://github.com/newstack-cloud/bluelink-provider-aws/commit/958fec951fb19b34fefd7e7b21ee4f0e79006bf0))

