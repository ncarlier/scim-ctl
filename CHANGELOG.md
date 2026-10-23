# [](https://github.com/ncarlier/scim-ctl/compare/v1.4.0...v) (2026-10-23)


### Features

* add Cursor-Based Pagination support ([1c1f0bf](https://github.com/ncarlier/scim-ctl/commit/1c1f0bfa6c75ee4da83e42ec313b21238d39660d))
* add env variable to define config location ([0f8b8c5](https://github.com/ncarlier/scim-ctl/commit/0f8b8c5e584c59caf046c7097d96b4995239b042))
* **auth:** add OIDC client AuthN with private key ([555a558](https://github.com/ncarlier/scim-ctl/commit/555a558bb2af156d12f23c1443668be1b0c1987a))
* **auth:** support encrypted PKCS[#8](https://github.com/ncarlier/scim-ctl/issues/8) file ([b8e6849](https://github.com/ncarlier/scim-ctl/commit/b8e68496507bb19f55a8654831eb46de1706b80c))
* **export:** diplay ETA or elapsed time ([7402149](https://github.com/ncarlier/scim-ctl/commit/7402149ecb165971ee60b2200ca328299d5e8c21))
* **export:** manage optional totalResults presence ([2aa578a](https://github.com/ncarlier/scim-ctl/commit/2aa578aae75da710926c65ef9ec9d3555c0b664f))
* make authentication optional ([3ea9dda](https://github.com/ncarlier/scim-ctl/commit/3ea9dda9270349eae30d4d6983cfba058c25d4d6))
* **search:** allow 0 items per page ([c69f880](https://github.com/ncarlier/scim-ctl/commit/c69f88092a959cd7df65d2686d3e4201b830e122))
* support excluded attributes parameter ([0a322e1](https://github.com/ncarlier/scim-ctl/commit/0a322e13ac0cfd29e86255389ea2b19fc089bf79))



# [1.3.0](https://github.com/ncarlier/scim-ctl/compare/v1.2.0...v1.3.0) (2026-09-08)


### Features

* add bulk-delete command ([a432722](https://github.com/ncarlier/scim-ctl/commit/a432722b0e8509dbda5ce0b4782a5aae331a5852))



# [1.2.0](https://github.com/ncarlier/scim-ctl/compare/v1.1.1...v1.2.0) (2026-08-31)


### Features

* add bulk-update command ([77feb04](https://github.com/ncarlier/scim-ctl/commit/77feb047c447189ae18b6faa64e79f4ac98706f6))
* print import progression ([235c70d](https://github.com/ncarlier/scim-ctl/commit/235c70d3ef82417328067ebcdb00885bd8194aca))



## [1.1.1](https://github.com/ncarlier/scim-ctl/compare/v1.1.0...v1.1.1) (2026-08-31)


### Features

* **export:** print export progression ([baee8b8](https://github.com/ncarlier/scim-ctl/commit/baee8b8904b5f363e8d0dad699eef2af52e85d3e))



# [1.1.0](https://github.com/ncarlier/scim-ctl/compare/v1.0.2...v1.1.0) (2026-08-25)


### Features

* add max-retries parameter ([e4ba9a4](https://github.com/ncarlier/scim-ctl/commit/e4ba9a4a522cf8063d9cb0b67c413dfcfaf373e1))
* add timeout parameter ([c124c71](https://github.com/ncarlier/scim-ctl/commit/c124c71e711d2d8e13eebba8849fe449406f2c26))



## [1.0.2](https://github.com/ncarlier/scim-ctl/compare/v1.0.1...v1.0.2) (2026-07-08)


### Features

* create install script ([da3f38a](https://github.com/ncarlier/scim-ctl/commit/da3f38ae68383f8b758075ec2af27e89241cce05))
* use JSON as standard STDOUT format ([8088b13](https://github.com/ncarlier/scim-ctl/commit/8088b137ddc1b9b628fc1898720a07f22802be7f))



## [1.0.1](https://github.com/ncarlier/scim-ctl/compare/v1.0.0...v1.0.1) (2026-07-01)



# [1.0.0](https://github.com/ncarlier/scim-ctl/compare/e8e6c0c2134c4a4a180857463e3410ad4b56de4c...v1.0.0) (2026-07-01)


### Bug Fixes

* **auth:** silent renew for access token ([7f6d93a](https://github.com/ncarlier/scim-ctl/commit/7f6d93ae6bad747f0d1fc2a71aef871cba775990))
* improve error management ([29c6e4c](https://github.com/ncarlier/scim-ctl/commit/29c6e4c2c04c722879df10f6cdf8877e9fb13877))


### Features

* add attributes optional parameter support ([16fae70](https://github.com/ncarlier/scim-ctl/commit/16fae7060a6913122f87fd92bcf8bd284a07c352))
* add client_credential grant type support ([4a52e45](https://github.com/ncarlier/scim-ctl/commit/4a52e4509f3e93e267c314b78770b1d0f80c0716))
* add export command ([f2d9e37](https://github.com/ncarlier/scim-ctl/commit/f2d9e37dcff72f38f7d25dd48111447f8189416a))
* add extra header ([34d7277](https://github.com/ncarlier/scim-ctl/commit/34d7277202c00198a59fe8b1078216aeb7e4b949))
* add global version flag ([8671329](https://github.com/ncarlier/scim-ctl/commit/86713299966b12891f2da76b2f1cb7366c70d5a8))
* add me command and improve HTTP client request ([f78be68](https://github.com/ncarlier/scim-ctl/commit/f78be6870c84b8d840dc2736e3245015d74ae4e9))
* add resource-type command ([e611db7](https://github.com/ncarlier/scim-ctl/commit/e611db7ccffc2d00a93057049d8d3f8f9264763e))
* add scripting sample ([2c39bb7](https://github.com/ncarlier/scim-ctl/commit/2c39bb76ec7124d2d6717bed0ba983977410d1d3))
* add search sort support ([001fd50](https://github.com/ncarlier/scim-ctl/commit/001fd5070ff80336cb575fcdc759d15d57c119ec))
* add update operation ([82a07ac](https://github.com/ncarlier/scim-ctl/commit/82a07ac514e025efb8b50514672d880048a319a1))
* allow cache dir configuration ([fdd28c0](https://github.com/ncarlier/scim-ctl/commit/fdd28c0c86ae056821e06a54aca47fd2c7f5bbe5))
* **auth:** add cache for access token ([e8e6c0c](https://github.com/ncarlier/scim-ctl/commit/e8e6c0c2134c4a4a180857463e3410ad4b56de4c))
* new extra header global flag ([3d2cc61](https://github.com/ncarlier/scim-ctl/commit/3d2cc613f93016dc85e4b1967936b85c30139df4))
* **search:** add full-text query support ([b90bcca](https://github.com/ncarlier/scim-ctl/commit/b90bccaffaeab37d47150ed8db5c2e6efbe97393))
* **search:** rename query parameter to filter ([f035f83](https://github.com/ncarlier/scim-ctl/commit/f035f83ed0d4093ff9342339b5d983a6ce726ca4))



