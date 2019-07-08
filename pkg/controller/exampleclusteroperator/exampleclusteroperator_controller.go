package exampleclusteroperator

import (
	"context"

	appv1alpha1 "github.com/djzager/example-cluster-operator/pkg/apis/app/v1alpha1"
	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	cov1helpers "github.com/openshift/library-go/pkg/config/clusteroperator/v1helpers"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_exampleclusteroperator")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ExampleClusterOperator Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileExampleClusterOperator{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("exampleclusteroperator-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ExampleClusterOperator
	err = c.Watch(&source.Kind{Type: &appv1alpha1.ExampleClusterOperator{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileExampleClusterOperator implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileExampleClusterOperator{}

// ReconcileExampleClusterOperator reconciles a ExampleClusterOperator object
type ReconcileExampleClusterOperator struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ExampleClusterOperator object and makes changes based on the state read
// and what is in the ExampleClusterOperator.Spec
func (r *ReconcileExampleClusterOperator) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ExampleClusterOperator")

	// Fetch the ExampleClusterOperator instance
	instance := &appv1alpha1.ExampleClusterOperator{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	for _, f := range []func(*appv1alpha1.ExampleClusterOperator, logr.Logger) error{
		r.syncAvailable,
		r.syncProgressing,
		r.syncDegraded,
		r.syncUpgradeable,
	} {
		err = f(instance, reqLogger)
		if err != nil {
			reqLogger.Info("Failed to sync cluster operator status")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileExampleClusterOperator) fetchClusterOperator(instance *appv1alpha1.ExampleClusterOperator) (*configv1.ClusterOperator, error) {
	found := &configv1.ClusterOperator{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: ""}, found)

	// If ClusterOperator CRD not present
	if meta.IsNoMatchError(err) {
		return nil, nil
	}
	if errors.IsNotFound(err) {
		clusterOperator := &configv1.ClusterOperator{
			ObjectMeta: metav1.ObjectMeta{
				Name: instance.Name,
			},
		}
		err = r.client.Create(context.TODO(), clusterOperator)
		if err != nil {
			log.Error(err, "Failed to create ClusterOperator")
			return nil, err
		}

		cov1helpers.SetStatusCondition(&clusterOperator.Status.Conditions, configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorAvailable, Status: configv1.ConditionFalse})
		cov1helpers.SetStatusCondition(&clusterOperator.Status.Conditions, configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorProgressing, Status: configv1.ConditionFalse})
		cov1helpers.SetStatusCondition(&clusterOperator.Status.Conditions, configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorDegraded, Status: configv1.ConditionFalse})
		cov1helpers.SetStatusCondition(&clusterOperator.Status.Conditions, configv1.ClusterOperatorStatusCondition{Type: configv1.OperatorUpgradeable, Status: configv1.ConditionFalse})

		// write the status
		return clusterOperator, r.client.Status().Update(context.TODO(), clusterOperator)
	}
	if err != nil {
		return nil, err
	}
	return found, nil
}

func (r *ReconcileExampleClusterOperator) updateClusterOperatorStatus(clusterOperator *configv1.ClusterOperator, status configv1.ClusterOperatorStatusCondition) error {
	existingCondition := cov1helpers.FindStatusCondition(clusterOperator.Status.Conditions, status.Type)
	if existingCondition.Status != status.Status {
		status.LastTransitionTime = metav1.Now()
	}

	cov1helpers.SetStatusCondition(&clusterOperator.Status.Conditions, status)
	return r.client.Status().Update(context.TODO(), clusterOperator)
}

func (r *ReconcileExampleClusterOperator) syncAvailable(instance *appv1alpha1.ExampleClusterOperator, reqLogger logr.Logger) error {
	if instance.Spec.OperatorAvailable == "" {
		reqLogger.Info("OperatorAvailable not set")
		return nil
	}
	available := configv1.ConditionStatus(instance.Spec.OperatorAvailable)

	co, err := r.fetchClusterOperator(instance)
	if err != nil {
		return err
	}
	if co == nil {
		reqLogger.Info("There is no ClusterOperator CRD")
		return nil
	}

	return r.updateClusterOperatorStatus(co, configv1.ClusterOperatorStatusCondition{
		Type:    configv1.OperatorAvailable,
		Status:  available,
		Message: "Example available message",
	})
}

func (r *ReconcileExampleClusterOperator) syncProgressing(instance *appv1alpha1.ExampleClusterOperator, reqLogger logr.Logger) error {
	if instance.Spec.OperatorProgressing == "" {
		reqLogger.Info("OperatorProgressing not set")
		return nil
	}
	progressing := configv1.ConditionStatus(instance.Spec.OperatorProgressing)

	co, err := r.fetchClusterOperator(instance)
	if err != nil {
		return err
	}
	if co == nil {
		reqLogger.Info("There is no ClusterOperator CRD")
		return nil
	}

	return r.updateClusterOperatorStatus(co, configv1.ClusterOperatorStatusCondition{
		Type:    configv1.OperatorProgressing,
		Status:  progressing,
		Message: "Example progressing message",
	})
}

func (r *ReconcileExampleClusterOperator) syncDegraded(instance *appv1alpha1.ExampleClusterOperator, reqLogger logr.Logger) error {
	if instance.Spec.OperatorDegraded == "" {
		reqLogger.Info("OperatorDegraded not set")
		return nil
	}
	degraded := configv1.ConditionStatus(instance.Spec.OperatorDegraded)

	co, err := r.fetchClusterOperator(instance)
	if err != nil {
		return err
	}
	if co == nil {
		reqLogger.Info("There is no ClusterOperator CRD")
		return nil
	}

	return r.updateClusterOperatorStatus(co, configv1.ClusterOperatorStatusCondition{
		Type:    configv1.OperatorDegraded,
		Status:  degraded,
		Message: "Example degraded message",
	})
}

func (r *ReconcileExampleClusterOperator) syncUpgradeable(instance *appv1alpha1.ExampleClusterOperator, reqLogger logr.Logger) error {
	if instance.Spec.OperatorUpgradeable == "" {
		reqLogger.Info("OperatorUpgradeable not set")
		return nil
	}
	upgradeable := configv1.ConditionStatus(instance.Spec.OperatorUpgradeable)

	co, err := r.fetchClusterOperator(instance)
	if err != nil {
		return err
	}
	if co == nil {
		reqLogger.Info("There is no ClusterOperator CRD")
		return nil
	}

	return r.updateClusterOperatorStatus(co, configv1.ClusterOperatorStatusCondition{
		Type:    configv1.OperatorUpgradeable,
		Status:  upgradeable,
		Message: "Example upgradeable message",
	})
}
