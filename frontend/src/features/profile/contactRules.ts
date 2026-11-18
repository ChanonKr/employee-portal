import { z } from 'zod'
import type { ContactKey } from './api'

// Mirrors backend/modules/profile/domain/contact.go so users see mistakes before submitting.

/** Mobile numbers are typed as digits only (the input drops anything else). */
export const MOBILE_DIGITS = 10
export const onlyDigits = (v: string) => v.replace(/\D/g, '').slice(0, MOBILE_DIGITS)

// Same pattern as the server: letters, digits and . _ % + - before the @ (no leading,
// trailing or double dots), then dot-separated domain labels and a TLD of 2+ letters.
const emailPattern =
  /^[a-z0-9_%+-]+(?:\.[a-z0-9_%+-]+)*@(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,63}$/i

export const contactSchemas: Record<ContactKey, z.ZodType<string, string>> = {
  mobile: z
    .string()
    .trim()
    .regex(/^[0-9]{10}$/, 'Enter a 10-digit mobile number, for example 0812345678'),
  email: z
    .string()
    .trim()
    .min(1, 'Enter your email address')
    .max(254, 'Use at most 254 characters')
    .regex(emailPattern, 'Enter a valid email address, like name@example.com'),
  address: z.string().trim().min(1, 'Enter your address').max(500, 'Use at most 500 characters'),
}
