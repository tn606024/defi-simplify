import { readFile } from 'node:fs/promises';

import * as AaveV3Base from '@aave-dao/aave-address-book/AaveV3Base';

const packageMetadataURL = new URL(
  './node_modules/@aave-dao/aave-address-book/package.json',
  import.meta.url,
);
const packageMetadata = JSON.parse(await readFile(packageMetadataURL, 'utf8'));
const source = JSON.parse(await readFile(new URL('./source.json', import.meta.url), 'utf8'));

function requireAddress(name, value) {
  if (typeof value !== 'string' || !/^0x[0-9a-fA-F]{40}$/.test(value)) {
    throw new Error(`${name} is not a valid EVM address`);
  }
  return value;
}

const issuerSources = {
  USDC: 'https://developers.circle.com/stablecoins/usdc-contract-addresses',
};

if (packageMetadata.name !== '@aave-dao/aave-address-book') {
  throw new Error(`unexpected package name ${packageMetadata.name}`);
}
if (typeof packageMetadata.version !== 'string' || packageMetadata.version === '') {
  throw new Error('Address Book package version is missing');
}
if (source.repository !== 'https://github.com/aave-dao/aave-address-book') {
  throw new Error(`unexpected source repository ${source.repository}`);
}
if (source.package !== packageMetadata.name || source.packageVersion !== packageMetadata.version) {
  throw new Error('pinned source metadata does not match the installed Address Book package');
}
if (source.release !== `v${source.packageVersion}`) {
  throw new Error('pinned Address Book release does not match its package version');
}
if (typeof source.commit !== 'string' || !/^[0-9a-fA-F]{40}$/.test(source.commit)) {
  throw new Error('pinned Address Book commit is missing or malformed');
}
if (source.export !== 'AaveV3Base') {
  throw new Error(`unexpected Address Book export ${source.export}`);
}
if (AaveV3Base.CHAIN_ID !== 8453) {
  throw new Error(`AaveV3Base has unexpected chain ID ${AaveV3Base.CHAIN_ID}`);
}

const exported = {
  packageName: packageMetadata.name,
  packageVersion: packageMetadata.version,
  gitHead: source.commit,
  export: source.export,
  chainId: AaveV3Base.CHAIN_ID,
  contracts: {
    poolAddressesProvider: requireAddress(
      'POOL_ADDRESSES_PROVIDER',
      AaveV3Base.POOL_ADDRESSES_PROVIDER,
    ),
    pool: requireAddress('POOL', AaveV3Base.POOL),
    aaveProtocolDataProvider: requireAddress(
      'AAVE_PROTOCOL_DATA_PROVIDER',
      AaveV3Base.AAVE_PROTOCOL_DATA_PROVIDER,
    ),
    wrappedTokenGateway: requireAddress('WETH_GATEWAY', AaveV3Base.WETH_GATEWAY),
  },
  assets: Object.entries(AaveV3Base.ASSETS).map(([key, asset]) => ({
    key,
    address: requireAddress(`${key}.UNDERLYING`, asset.UNDERLYING),
    ...(issuerSources[key] ? { issuerSource: issuerSources[key] } : {}),
  })),
};

process.stdout.write(`${JSON.stringify(exported)}\n`);
