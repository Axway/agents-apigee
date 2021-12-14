const { execSync } = require("child_process");

const packagename = "models";
const pathname = "apigee";
const group = "models";

execSync(
	`openapi-generator generate -g go -i ./specs/apis.yaml --package-name ${packagename} --output pkg/${pathname}/${group}/ --global-property modelDocs=false --global-property models --global-property apiDocs=false`
);
execSync(
	`openapi-generator generate -g go -i ./specs/sharedflows.yaml --package-name ${packagename} --output pkg/${pathname}/${group}/ --global-property modelDocs=false --global-property models --global-property apiDocs=false`
);
execSync(
	`openapi-generator generate -g go -i ./specs/virtualhosts.yaml --package-name ${packagename} --output pkg/${pathname}/${group}/ --global-property modelDocs=false --global-property models --global-property apiDocs=false`
);
execSync(
	`openapi-generator generate -g go -i ./specs/apiproducts.yaml --package-name ${packagename} --output pkg/${pathname}/${group}/ --global-property modelDocs=false --global-property models --global-property apiDocs=false`
);
execSync(
	`openapi-generator generate -g go -i ./specs/developers.yaml --package-name ${packagename} --output pkg/${pathname}/${group}/ --global-property modelDocs=false --global-property models --global-property apiDocs=false`
);
execSync(
	`openapi-generator generate -g go -i ./specs/developer_apps.yaml --package-name ${packagename} --output pkg/${pathname}/${group}/ --global-property modelDocs=false --global-property models --global-property apiDocs=false`
);
