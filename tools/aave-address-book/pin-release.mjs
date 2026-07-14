import { readFile, writeFile } from 'node:fs/promises';

const [version, commit] = process.argv.slice(2);
if (typeof version !== 'string' || !/^\d+\.\d+\.\d+(?:[-+].+)?$/.test(version)) {
  throw new Error('usage: node pin-release.mjs <package-version> <40-character-git-commit>');
}
if (typeof commit !== 'string' || !/^[0-9a-fA-F]{40}$/.test(commit)) {
  throw new Error('usage: node pin-release.mjs <package-version> <40-character-git-commit>');
}

const packageURL = new URL('./package.json', import.meta.url);
const packageJSON = JSON.parse(await readFile(packageURL, 'utf8'));
packageJSON.dependencies['@aave-dao/aave-address-book'] = version;

const source = {
  repository: 'https://github.com/aave-dao/aave-address-book',
  package: '@aave-dao/aave-address-book',
  packageVersion: version,
  release: `v${version}`,
  commit: commit.toLowerCase(),
  export: 'AaveV3Base',
};

await writeFile(packageURL, `${JSON.stringify(packageJSON, null, 2)}\n`);
await writeFile(new URL('./source.json', import.meta.url), `${JSON.stringify(source, null, 2)}\n`);
