const { execSync } = require("child_process");

const packagename = "models";
const pathname = "apigee";
const group = "models";

// Receiving a warning in the cli...
// [DEPRECATED] -D arguments after 'generate' are application arguments and not Java System Properties,
// please consider changing to --global-property, apply your system properties to JAVA_OPTS,
// or move the -D arguments before the jar option.
// changed -Dmodels to: --global-property models
execSync(
	`openapi-generator generate -g go -i ./specs/apis.yaml --package-name ${packagename} --output pkg/${pathname}/${group}/ --global-property modelDocs=false --global-property models --global-property apiDocs=false`
);
execSync(
	`openapi-generator generate -g go -i ./specs/deployments.yaml --package-name ${packagename} --output pkg/${pathname}/${group}/  --skip-validate-spec --global-property modelDocs=false --global-property models --global-property apiDocs=false`
);
execSync(
	`openapi-generator generate -g go -i ./specs/environments.yaml --package-name ${packagename} --output pkg/${pathname}/${group}/ --global-property modelDocs=false --global-property models --global-property apiDocs=false`
);
execSync(
	`openapi-generator generate -g go -i ./specs/virtualhosts.yaml --package-name ${packagename} --output pkg/${pathname}/${group}/ --global-property modelDocs=false --global-property models --global-property apiDocs=false`
);
execSync(
	`openapi-generator generate -g go -i ./specs/apiproducts.yaml --package-name ${packagename} --output pkg/${pathname}/${group}/ --global-property modelDocs=false --global-property models --global-property apiDocs=false`
);
