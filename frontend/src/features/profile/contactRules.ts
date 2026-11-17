import { z } from 'zod'
import type { ContactKey } from './api'

// Mirrors entity/contact.go so users see mistakes before submitting.
export const contactSchemas: Record<ContactKey, z.ZodType<string, string>> = {
  mobile: z
    .string()
    .transform((v) => v.trim().replace(/[\s\-()]/g, ''))
    .pipe(z.string().regex(/^\+?[0-9]{9,15}$/, 'Enter 9–15 digits, for example 081 234 5678')),
  email: z
    .string()
    .trim()
    .pipe(z.email('Enter an email address like name@example.com').max(254, 'Use at most 254 characters')),
  address: z.string().trim().min(1, 'Enter your address').max(500, 'Use at most 500 characters'),
}
