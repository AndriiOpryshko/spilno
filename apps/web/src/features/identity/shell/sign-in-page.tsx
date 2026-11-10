import { BrandMark } from '@/components/app-shell';
import { Button } from '@/components/ui/button';
import { uk } from '@/i18n/uk';

export function SignInPage() {
  return (
    <div className="mx-auto flex max-w-sm flex-col items-center gap-4 px-4 py-24 text-center">
      <BrandMark />
      <h1 className="text-2xl font-semibold">{uk.association.signInTitle}</h1>
      <Button asChild>
        <a href="/identity/sign-in">{uk.association.signIn}</a>
      </Button>
    </div>
  );
}
