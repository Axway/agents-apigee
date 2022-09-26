const fs = require("fs");
const { execSync } = require("child_process");

const outDir = process.env.OUTDIR;
const packagename = "models";

const specdir = "../specs";

fs.readdir(specdir, (err, files) => {
	if (err) console.log(err);
	else {
		files.forEach((file) => {
			console.log(`Generating models from spec: ${file}`);
			execSync(
				`openapi-generator-cli generate -g go -i ${specdir}/${file} --package-name ${packagename} --output ${outDir} --global-property modelDocs=false --global-property models --global-property apiDocs=false`
			);
		});
	}
});
