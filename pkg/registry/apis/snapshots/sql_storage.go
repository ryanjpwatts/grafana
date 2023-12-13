package snapshots

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/internalversion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/registry/rest"

	"github.com/grafana/grafana/pkg/apis/snapshots/v0alpha1"
	"github.com/grafana/grafana/pkg/infra/appcontext"
	"github.com/grafana/grafana/pkg/services/dashboardsnapshots"
	"github.com/grafana/grafana/pkg/services/grafana-apiserver/endpoints/request"
)

var (
	_ rest.Scoper               = (*legacyStorage)(nil)
	_ rest.SingularNameProvider = (*legacyStorage)(nil)
	_ rest.Getter               = (*legacyStorage)(nil)
	_ rest.Lister               = (*legacyStorage)(nil)
	_ rest.Storage              = (*legacyStorage)(nil)
	_ rest.GracefulDeleter      = (*legacyStorage)(nil)
)

type legacyStorage struct {
	service        dashboardsnapshots.Service
	namespacer     request.NamespaceMapper
	tableConverter rest.TableConvertor
	options        sharingOptionsGetter
}

func (s *legacyStorage) New() runtime.Object {
	return resourceInfo.NewFunc()
}

func (s *legacyStorage) Destroy() {}

func (s *legacyStorage) NamespaceScoped() bool {
	return true // namespace == org
}

func (s *legacyStorage) GetSingularName() string {
	return resourceInfo.GetSingularName()
}

func (s *legacyStorage) NewList() runtime.Object {
	return resourceInfo.NewListFunc()
}

func (s *legacyStorage) checkEnabled(ns string) error {
	opts, err := s.options(ns)
	if err != nil {
		return err
	}
	if !opts.Spec.SnapshotsEnabled {
		return fmt.Errorf("snapshots not enabled")
	}
	return nil
}

func (s *legacyStorage) ConvertToTable(ctx context.Context, object runtime.Object, tableOptions runtime.Object) (*metav1.Table, error) {
	return s.tableConverter.ConvertToTable(ctx, object, tableOptions)
}

func (s *legacyStorage) List(ctx context.Context, options *internalversion.ListOptions) (runtime.Object, error) {
	// TODO: handle fetching all available orgs when no namespace is specified
	// To test: kubectl get playlists --all-namespaces
	info, err := request.NamespaceInfoFrom(ctx, true)
	if err == nil {
		err = s.checkEnabled(info.Value)
	}
	if err != nil {
		return nil, err
	}

	user, err := appcontext.User(ctx)
	if err != nil {
		return nil, err
	}

	limit := 100
	if options.Limit > 0 {
		limit = int(options.Limit)
	}
	res, err := s.service.SearchDashboardSnapshots(ctx, &dashboardsnapshots.GetDashboardSnapshotsQuery{
		OrgID:        info.OrgID,
		SignedInUser: user,
	})
	if err != nil {
		return nil, err
	}

	list := &v0alpha1.DashboardSnapshotList{}
	for _, v := range res {
		list.Items = append(list.Items, *convertDTOToSnapshot(v, s.namespacer))
	}
	if len(list.Items) == limit {
		list.Continue = "<more>" // TODO?
	}
	return list, nil
}

func (s *legacyStorage) Get(ctx context.Context, name string, options *metav1.GetOptions) (runtime.Object, error) {
	info, err := request.NamespaceInfoFrom(ctx, true)
	if err == nil {
		err = s.checkEnabled(info.Value)
	}
	if err != nil {
		return nil, err
	}

	v, err := s.service.GetDashboardSnapshot(ctx, &dashboardsnapshots.GetDashboardSnapshotQuery{
		Key: name,
	})
	if err != nil || v == nil {
		// if errors.Is(err, playlistsvc.ErrPlaylistNotFound) || err == nil {
		// 	err = k8serrors.NewNotFound(s.SingularQualifiedResource, name)
		// }
		return nil, err
	}

	return convertSnapshotToK8sResource(v, s.namespacer), nil
}

// GracefulDeleter
func (s *legacyStorage) Delete(ctx context.Context, name string, deleteValidation rest.ValidateObjectFunc, options *metav1.DeleteOptions) (runtime.Object, bool, error) {
	snap, err := s.service.GetDashboardSnapshot(ctx, &dashboardsnapshots.GetDashboardSnapshotQuery{
		Key: name,
	})
	if err != nil || snap == nil {
		return nil, false, err
	}

	// Delete the external one first
	if snap.ExternalDeleteURL != "" {
		err := dashboardsnapshots.DeleteExternalDashboardSnapshot(snap.ExternalDeleteURL)
		if err != nil {
			return nil, false, err
		}
	}

	err = s.service.DeleteDashboardSnapshot(ctx, &dashboardsnapshots.DeleteDashboardSnapshotCommand{
		DeleteKey: snap.DeleteKey,
	})
	if err != nil {
		return nil, false, err
	}
	return nil, true, nil
}
