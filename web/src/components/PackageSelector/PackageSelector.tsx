import type { PackageSummary } from '../../types';
import { PackageSelectorParts } from './parts';
import './styles/package-selector.css';

interface PackageSelectorProps {
  packages: PackageSummary[];
  selectedPackage: string;
  selectedVersion: string;
  onPackageChange: (value: string) => void;
  onVersionChange: (value: string) => void;
}

export function PackageSelector({
  packages,
  selectedPackage,
  selectedVersion,
  onPackageChange,
  onVersionChange,
}: PackageSelectorProps) {
  const selected = packages.find((pkg) => pkg.name === selectedPackage);
  const versions = selected?.versions ?? [];

  if (packages.length === 0) {
    return null;
  }

  return (
    <div data-part={PackageSelectorParts.root}>
      <label data-part={PackageSelectorParts.row}>
        <span data-part={PackageSelectorParts.label}>Package</span>
        <select
          data-part={PackageSelectorParts.select}
          value={selectedPackage}
          onChange={(event) => onPackageChange(event.target.value)}
        >
          {packages.map((pkg) => (
            <option key={pkg.name} value={pkg.name}>
              {pkg.displayName || pkg.name}
            </option>
          ))}
        </select>
      </label>
      {versions.length > 0 && (
        <label data-part={PackageSelectorParts.row}>
          <span data-part={PackageSelectorParts.label}>Version</span>
          <select
            data-part={PackageSelectorParts.select}
            value={selectedVersion}
            onChange={(event) => onVersionChange(event.target.value)}
          >
            {versions.map((version) => (
              <option key={version} value={version}>
                {version}
              </option>
            ))}
          </select>
        </label>
      )}
    </div>
  );
}
