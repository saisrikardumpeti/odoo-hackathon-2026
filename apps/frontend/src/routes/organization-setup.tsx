import { createFileRoute, redirect } from '@tanstack/react-router';
import { useAuthStore } from '#/lib/stores/authStore';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '#/components/ui/tabs';
import { Building2, Tags, Users } from 'lucide-react';
import { DepartmentTab } from '#/components/organization-setup/DepartmentTab';
import { CategoryTab } from '#/components/organization-setup/CategoryTab';
import { EmployeeTab } from '#/components/organization-setup/EmployeeTab';

export const Route = createFileRoute('/organization-setup')({
  beforeLoad: () => {
    const { isAuthenticated, employee } = useAuthStore.getState();
    if (!isAuthenticated) {
      throw redirect({ to: '/auth/login' });
    }
    if (employee?.role !== 'Admin') {
      throw redirect({ to: '/' });
    }
  },
  component: OrganizationSetupPage,
});

function OrganizationSetupPage() {
  return (
    <div className="p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Organization Setup</h1>
        <p className="text-muted-foreground text-sm">Manage departments, asset categories, and employee directory.</p>
      </div>

      <Tabs defaultValue="departments">
        <TabsList>
          <TabsTrigger value="departments">
            <Building2 className="mr-2 size-4" /> Departments
          </TabsTrigger>
          <TabsTrigger value="categories">
            <Tags className="mr-2 size-4" /> Categories
          </TabsTrigger>
          <TabsTrigger value="employees">
            <Users className="mr-2 size-4" /> Employee Directory
          </TabsTrigger>
        </TabsList>
        <TabsContent value="departments" className="mt-4">
          <DepartmentTab />
        </TabsContent>
        <TabsContent value="categories" className="mt-4">
          <CategoryTab />
        </TabsContent>
        <TabsContent value="employees" className="mt-4">
          <EmployeeTab />
        </TabsContent>
      </Tabs>
    </div>
  );
}