export interface CountryCode {
  name: string;
  dialCode: string;
}

// A short, common list rather than all ~240 countries — kept easy to scan
// in a dropdown. India first since it's the primary market.
export const COUNTRY_CODES: CountryCode[] = [
  { name: 'India', dialCode: '+91' },
  { name: 'United States', dialCode: '+1' },
  { name: 'United Kingdom', dialCode: '+44' },
  { name: 'UAE', dialCode: '+971' },
  { name: 'Saudi Arabia', dialCode: '+966' },
  { name: 'Qatar', dialCode: '+974' },
  { name: 'Kuwait', dialCode: '+965' },
  { name: 'Oman', dialCode: '+968' },
  { name: 'Bahrain', dialCode: '+973' },
  { name: 'Singapore', dialCode: '+65' },
  { name: 'Malaysia', dialCode: '+60' },
  { name: 'Australia', dialCode: '+61' },
  { name: 'Canada', dialCode: '+1' },
  { name: 'Sri Lanka', dialCode: '+94' },
  { name: 'Bangladesh', dialCode: '+880' },
  { name: 'Pakistan', dialCode: '+92' },
  { name: 'Nepal', dialCode: '+977' },
  { name: 'Germany', dialCode: '+49' },
  { name: 'France', dialCode: '+33' },
  { name: 'South Africa', dialCode: '+27' },
];

export const DEFAULT_COUNTRY_DIAL_CODE = '+91';
