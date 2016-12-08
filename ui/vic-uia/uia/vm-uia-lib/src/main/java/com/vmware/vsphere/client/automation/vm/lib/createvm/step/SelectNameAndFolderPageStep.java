/* Copyright 2014 VMware, Inc. All rights reserved. -- VMware Confidential */
package com.vmware.vsphere.client.automation.vm.lib.createvm.step;

import com.vmware.client.automation.common.view.BaseView;
import com.vmware.vsphere.client.automation.vm.lib.createvm.view.SelectNameAndFolderPage;

public class SelectNameAndFolderPageStep extends CreateVmFlowStep {

   @Override
   public void execute() throws Exception {
      SelectNameAndFolderPage selectPage = new SelectNameAndFolderPage();

      selectPage.waitForDialogToLoad();
      // Set VM name
      selectPage.setVmName(createVmSpec.name.get());
      selectPage.waitApplySavedDataProgressBar();
      new BaseView().waitForPageToRefresh();
      // Click on Next and verify that next page is loaded
      boolean isNextButtonClicked = selectPage.gotoNextPage();
      verifyFatal(isNextButtonClicked, "Verify the next button is clicked!");
   }
}